/*
Copyright 2021 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package reparentutil

import (
	"context"
	"time"

	replicationdatapb "vitess.io/vitess/go/vt/proto/replicationdata"

	"vitess.io/vitess/go/vt/topo/topoproto"

	"vitess.io/vitess/go/vt/proto/vtrpc"

	topodatapb "vitess.io/vitess/go/vt/proto/topodata"

	"google.golang.org/protobuf/proto"

	"vitess.io/vitess/go/vt/topo"
	"vitess.io/vitess/go/vt/vterrors"

	"vitess.io/vitess/go/event"

	"vitess.io/vitess/go/vt/logutil"
	logutilpb "vitess.io/vitess/go/vt/proto/logutil"
	"vitess.io/vitess/go/vt/topotools/events"
	"vitess.io/vitess/go/vt/vttablet/tmclient"
)

// EmergencyReparenter performs EmergencyReparentShard operations.
type EmergencyReparenter struct {
	ts     *topo.Server
	tmc    tmclient.TabletManagerClient
	logger logutil.Logger
}

// NewEmergencyReparenter returns a new EmergencyReparenter object, ready to
// perform EmergencyReparentShard operations using the given topo.Server,
// TabletManagerClient, and logger.
//
// Providing a nil logger instance is allowed.
func NewEmergencyReparenter(ts *topo.Server, tmc tmclient.TabletManagerClient, logger logutil.Logger) *EmergencyReparenter {
	erp := EmergencyReparenter{
		ts:     ts,
		tmc:    tmc,
		logger: logger,
	}

	if erp.logger == nil {
		// Create a no-op logger so we can call functions on er.logger without
		// needed to constantly check for non-nil.
		erp.logger = logutil.NewCallbackLogger(func(*logutilpb.Event) {})
	}

	return &erp
}

// ReparentShard performs the EmergencyReparentShard operation on the given
// keyspace and shard.
func (erp *EmergencyReparenter) ReparentShard(ctx context.Context, keyspace, shard string, reparentFunctions ReparentFunctions) (*events.Reparent, error) {
	// First step is to lock the shard for the given operation
	ctx, unlock, err := reparentFunctions.LockShard(ctx, erp.logger, erp.ts, keyspace, shard)
	if err != nil {
		return nil, err
	}
	// defer the unlock-shard function
	defer unlock(&err)

	// dispatch success or failure of ERS
	ev := &events.Reparent{}
	defer func() {
		switch err {
		case nil:
			event.DispatchUpdate(ev, "finished EmergencyReparentShard")
		default:
			event.DispatchUpdate(ev, "failed EmergencyReparentShard: "+err.Error())
		}
	}()

	// run ERS with shard already locked
	err = erp.reparentShardLocked(ctx, ev, keyspace, shard, reparentFunctions)

	reparentFunctions.PostERSCompletionHook(ctx, ev, erp.logger, erp.tmc)

	return ev, err
}

// reparentShardLocked performs Emergency Reparent Shard operation assuming that the shard is already locked
func (erp *EmergencyReparenter) reparentShardLocked(ctx context.Context, ev *events.Reparent, keyspace, shard string, reparentFunctions ReparentFunctions) error {
	// check whether ERS is required or it has been fixed via an external agent
	if reparentFunctions.CheckIfFixed() {
		return nil
	}

	// get the shard information from the topology server
	shardInfo, err := erp.ts.GetShard(ctx, keyspace, shard)
	if err != nil {
		return err
	}
	ev.ShardInfo = *shardInfo

	// get the previous primary according to the topology server
	var prevPrimary *topodatapb.Tablet
	if shardInfo.PrimaryAlias != nil {
		prevPrimaryInfo, err := erp.ts.GetTablet(ctx, shardInfo.PrimaryAlias)
		if err != nil {
			return err
		}
		prevPrimary = prevPrimaryInfo.Tablet
	}

	// run the pre recovery processes
	if err := reparentFunctions.PreRecoveryProcesses(ctx); err != nil {
		return err
	}

	if err := reparentFunctions.CheckPrimaryRecoveryType(); err != nil {
		return err
	}

	event.DispatchUpdate(ev, "reading all tablets")
	tabletMap, err := erp.ts.GetTabletMapForShard(ctx, keyspace, shard)
	if err != nil {
		return vterrors.Wrapf(err, "failed to get tablet map for %v/%v: %v", keyspace, shard, err)
	}

	statusMap, primaryStatusMap, err := StopReplicationAndBuildStatusMaps(ctx, erp.tmc, ev, tabletMap, reparentFunctions.GetWaitReplicasTimeout(), reparentFunctions.GetIgnoreReplicas(), erp.logger)
	if err != nil {
		return vterrors.Wrapf(err, "failed to stop replication and build status maps: %v", err)
	}

	if err := topo.CheckShardLocked(ctx, keyspace, shard); err != nil {
		return vterrors.Wrapf(err, "lost topology lock, aborting: %v", err)
	}

	validCandidates, err := FindValidEmergencyReparentCandidates(statusMap, primaryStatusMap)
	if err != nil {
		return err
	}
	validCandidates, err = reparentFunctions.RestrictValidCandidates(validCandidates, tabletMap)
	if err != nil {
		return err
	} else if len(validCandidates) == 0 {
		return vterrors.Errorf(vtrpc.Code_FAILED_PRECONDITION, "no valid candidates for emergency reparent")
	}

	// Wait for all candidates to apply relay logs
	if err := waitForAllRelayLogsToApply(ctx, erp.logger, erp.tmc, validCandidates, tabletMap, statusMap, reparentFunctions.GetWaitForRelayLogsTimeout()); err != nil {
		err = reparentFunctions.HandleRelayLogFailure(err)
		if err != nil {
			return err
		}
	}

	var newPrimary *topodatapb.Tablet
	newPrimary, tabletMap, err = reparentFunctions.FindPrimaryCandidates(ctx, erp.logger, erp.tmc, validCandidates, tabletMap)
	if err != nil {
		return err
	}

	isIdeal := reparentFunctions.PromotedReplicaIsIdeal(newPrimary, prevPrimary, tabletMap, validCandidates)

	// TODO := LockAction and RP
	validReplacementCandidates, err := promotePrimaryCandidateAndStartReplication(ctx, erp.tmc, erp.ts, ev, erp.logger, newPrimary, "", "", tabletMap, statusMap, reparentFunctions, isIdeal, true)
	if err != nil {
		return err
	}

	// Check (again) we still have the topology lock.
	if err := topo.CheckShardLocked(ctx, keyspace, shard); err != nil {
		return vterrors.Wrapf(err, "lost topology lock, aborting: %v", err)
	}

	betterCandidate := newPrimary
	if !isIdeal {
		betterCandidate = reparentFunctions.GetBetterCandidate(newPrimary, prevPrimary, validReplacementCandidates, tabletMap)
	}

	if !topoproto.TabletAliasEqual(betterCandidate.Alias, newPrimary.Alias) {
		err = replaceWithBetterCandidate(ctx, erp.tmc, erp.ts, ev, erp.logger, newPrimary, betterCandidate, "", "", tabletMap, statusMap, reparentFunctions)
		if err != nil {
			return err
		}
		newPrimary = betterCandidate
	}

	errInPromotion := reparentFunctions.CheckIfNeedToOverridePromotion(newPrimary)
	if errInPromotion != nil {
		erp.logger.Errorf("have to override promotion because of constraint failure - %v", errInPromotion)
		newPrimary, err = erp.undoPromotion(ctx, erp.ts, ev, keyspace, shard, prevPrimary, "", "", tabletMap, statusMap, reparentFunctions)
		if err != nil {
			return err
		}
		if newPrimary == nil {
			return vterrors.Errorf(vtrpc.Code_ABORTED, "could not undo promotion")
		}
	}

	_, err = erp.tmc.PromoteReplica(ctx, newPrimary)
	if err != nil {
		return vterrors.Wrapf(err, "primary-elect tablet %v failed to be upgraded to primary: %v", newPrimary.Alias, err)
	}
	reparentFunctions.PostTabletChangeHook(newPrimary)

	ev.NewPrimary = proto.Clone(newPrimary).(*topodatapb.Tablet)
	return errInPromotion
}

func (erp *EmergencyReparenter) undoPromotion(ctx context.Context, ts *topo.Server, ev *events.Reparent, keyspace, shard string, prevPrimary *topodatapb.Tablet,
	lockAction, rp string, tabletMap map[string]*topo.TabletInfo, statusMap map[string]*replicationdatapb.StopReplicationStatus, reparentFunctions ReparentFunctions) (*topodatapb.Tablet, error) {
	var primaryAlias *topodatapb.TabletAlias

	if prevPrimary != nil {
		_, err := promotePrimaryCandidateAndStartReplication(ctx, erp.tmc, ts, ev, erp.logger, prevPrimary, lockAction, rp, tabletMap, statusMap, reparentFunctions, true, false)
		if err == nil {
			return prevPrimary, nil
		}
		erp.logger.Errorf("error in undoing promotion - %v", err)
		primaryAlias = prevPrimary.Alias
	}

	newTerm := time.Now()
	_, err := ts.UpdateShardFields(ctx, keyspace, shard, func(si *topo.ShardInfo) error {
		if proto.Equal(si.PrimaryAlias, primaryAlias) {
			return nil
		}
		si.PrimaryAlias = primaryAlias
		si.PrimaryTermStartTime = logutil.TimeToProto(newTerm)
		return nil
	})
	return prevPrimary, err
}
