# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: binlogdata.proto

import sys
_b=sys.version_info[0]<3 and (lambda x:x) or (lambda x:x.encode('latin1'))
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
from google.protobuf import descriptor_pb2
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


import query_pb2 as query__pb2
import topodata_pb2 as topodata__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='binlogdata.proto',
  package='binlogdata',
  syntax='proto3',
  serialized_pb=_b('\n\x10\x62inlogdata.proto\x12\nbinlogdata\x1a\x0bquery.proto\x1a\x0etopodata.proto\"7\n\x07\x43harset\x12\x0e\n\x06\x63lient\x18\x01 \x01(\x05\x12\x0c\n\x04\x63onn\x18\x02 \x01(\x05\x12\x0e\n\x06server\x18\x03 \x01(\x05\"\xb5\x03\n\x11\x42inlogTransaction\x12;\n\nstatements\x18\x01 \x03(\x0b\x32\'.binlogdata.BinlogTransaction.Statement\x12&\n\x0b\x65vent_token\x18\x04 \x01(\x0b\x32\x11.query.EventToken\x1a\xae\x02\n\tStatement\x12\x42\n\x08\x63\x61tegory\x18\x01 \x01(\x0e\x32\x30.binlogdata.BinlogTransaction.Statement.Category\x12$\n\x07\x63harset\x18\x02 \x01(\x0b\x32\x13.binlogdata.Charset\x12\x0b\n\x03sql\x18\x03 \x01(\x0c\"\xa9\x01\n\x08\x43\x61tegory\x12\x13\n\x0f\x42L_UNRECOGNIZED\x10\x00\x12\x0c\n\x08\x42L_BEGIN\x10\x01\x12\r\n\tBL_COMMIT\x10\x02\x12\x0f\n\x0b\x42L_ROLLBACK\x10\x03\x12\x15\n\x11\x42L_DML_DEPRECATED\x10\x04\x12\n\n\x06\x42L_DDL\x10\x05\x12\n\n\x06\x42L_SET\x10\x06\x12\r\n\tBL_INSERT\x10\x07\x12\r\n\tBL_UPDATE\x10\x08\x12\r\n\tBL_DELETE\x10\tJ\x04\x08\x02\x10\x03J\x04\x08\x03\x10\x04\"v\n\x15StreamKeyRangeRequest\x12\x10\n\x08position\x18\x01 \x01(\t\x12%\n\tkey_range\x18\x02 \x01(\x0b\x32\x12.topodata.KeyRange\x12$\n\x07\x63harset\x18\x03 \x01(\x0b\x32\x13.binlogdata.Charset\"S\n\x16StreamKeyRangeResponse\x12\x39\n\x12\x62inlog_transaction\x18\x01 \x01(\x0b\x32\x1d.binlogdata.BinlogTransaction\"]\n\x13StreamTablesRequest\x12\x10\n\x08position\x18\x01 \x01(\t\x12\x0e\n\x06tables\x18\x02 \x03(\t\x12$\n\x07\x63harset\x18\x03 \x01(\x0b\x32\x13.binlogdata.Charset\"Q\n\x14StreamTablesResponse\x12\x39\n\x12\x62inlog_transaction\x18\x01 \x01(\x0b\x32\x1d.binlogdata.BinlogTransaction\"E\n\x0c\x42inlogFilter\x12%\n\tkey_range\x18\x01 \x01(\x0b\x32\x12.topodata.KeyRange\x12\x0e\n\x06tables\x18\x02 \x03(\tB)Z\'vitess.io/vitess/go/vt/proto/binlogdatab\x06proto3')
  ,
  dependencies=[query__pb2.DESCRIPTOR,topodata__pb2.DESCRIPTOR,])



_BINLOGTRANSACTION_STATEMENT_CATEGORY = _descriptor.EnumDescriptor(
  name='Category',
  full_name='binlogdata.BinlogTransaction.Statement.Category',
  filename=None,
  file=DESCRIPTOR,
  values=[
    _descriptor.EnumValueDescriptor(
      name='BL_UNRECOGNIZED', index=0, number=0,
      options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='BL_BEGIN', index=1, number=1,
      options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='BL_COMMIT', index=2, number=2,
      options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='BL_ROLLBACK', index=3, number=3,
      options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='BL_DML_DEPRECATED', index=4, number=4,
      options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='BL_DDL', index=5, number=5,
      options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='BL_SET', index=6, number=6,
      options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='BL_INSERT', index=7, number=7,
      options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='BL_UPDATE', index=8, number=8,
      options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='BL_DELETE', index=9, number=9,
      options=None,
      type=None),
  ],
  containing_type=None,
  options=None,
  serialized_start=375,
  serialized_end=544,
)
_sym_db.RegisterEnumDescriptor(_BINLOGTRANSACTION_STATEMENT_CATEGORY)


_CHARSET = _descriptor.Descriptor(
  name='Charset',
  full_name='binlogdata.Charset',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='client', full_name='binlogdata.Charset.client', index=0,
      number=1, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='conn', full_name='binlogdata.Charset.conn', index=1,
      number=2, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='server', full_name='binlogdata.Charset.server', index=2,
      number=3, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=61,
  serialized_end=116,
)


_BINLOGTRANSACTION_STATEMENT = _descriptor.Descriptor(
  name='Statement',
  full_name='binlogdata.BinlogTransaction.Statement',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='category', full_name='binlogdata.BinlogTransaction.Statement.category', index=0,
      number=1, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='charset', full_name='binlogdata.BinlogTransaction.Statement.charset', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='sql', full_name='binlogdata.BinlogTransaction.Statement.sql', index=2,
      number=3, type=12, cpp_type=9, label=1,
      has_default_value=False, default_value=_b(""),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
    _BINLOGTRANSACTION_STATEMENT_CATEGORY,
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=242,
  serialized_end=544,
)

_BINLOGTRANSACTION = _descriptor.Descriptor(
  name='BinlogTransaction',
  full_name='binlogdata.BinlogTransaction',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='statements', full_name='binlogdata.BinlogTransaction.statements', index=0,
      number=1, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='event_token', full_name='binlogdata.BinlogTransaction.event_token', index=1,
      number=4, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[_BINLOGTRANSACTION_STATEMENT, ],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=119,
  serialized_end=556,
)


_STREAMKEYRANGEREQUEST = _descriptor.Descriptor(
  name='StreamKeyRangeRequest',
  full_name='binlogdata.StreamKeyRangeRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='position', full_name='binlogdata.StreamKeyRangeRequest.position', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='key_range', full_name='binlogdata.StreamKeyRangeRequest.key_range', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='charset', full_name='binlogdata.StreamKeyRangeRequest.charset', index=2,
      number=3, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=558,
  serialized_end=676,
)


_STREAMKEYRANGERESPONSE = _descriptor.Descriptor(
  name='StreamKeyRangeResponse',
  full_name='binlogdata.StreamKeyRangeResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='binlog_transaction', full_name='binlogdata.StreamKeyRangeResponse.binlog_transaction', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=678,
  serialized_end=761,
)


_STREAMTABLESREQUEST = _descriptor.Descriptor(
  name='StreamTablesRequest',
  full_name='binlogdata.StreamTablesRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='position', full_name='binlogdata.StreamTablesRequest.position', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='tables', full_name='binlogdata.StreamTablesRequest.tables', index=1,
      number=2, type=9, cpp_type=9, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='charset', full_name='binlogdata.StreamTablesRequest.charset', index=2,
      number=3, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=763,
  serialized_end=856,
)


_STREAMTABLESRESPONSE = _descriptor.Descriptor(
  name='StreamTablesResponse',
  full_name='binlogdata.StreamTablesResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='binlog_transaction', full_name='binlogdata.StreamTablesResponse.binlog_transaction', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=858,
  serialized_end=939,
)


_BINLOGFILTER = _descriptor.Descriptor(
  name='BinlogFilter',
  full_name='binlogdata.BinlogFilter',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='key_range', full_name='binlogdata.BinlogFilter.key_range', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='tables', full_name='binlogdata.BinlogFilter.tables', index=1,
      number=2, type=9, cpp_type=9, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=941,
  serialized_end=1010,
)

_BINLOGTRANSACTION_STATEMENT.fields_by_name['category'].enum_type = _BINLOGTRANSACTION_STATEMENT_CATEGORY
_BINLOGTRANSACTION_STATEMENT.fields_by_name['charset'].message_type = _CHARSET
_BINLOGTRANSACTION_STATEMENT.containing_type = _BINLOGTRANSACTION
_BINLOGTRANSACTION_STATEMENT_CATEGORY.containing_type = _BINLOGTRANSACTION_STATEMENT
_BINLOGTRANSACTION.fields_by_name['statements'].message_type = _BINLOGTRANSACTION_STATEMENT
_BINLOGTRANSACTION.fields_by_name['event_token'].message_type = query__pb2._EVENTTOKEN
_STREAMKEYRANGEREQUEST.fields_by_name['key_range'].message_type = topodata__pb2._KEYRANGE
_STREAMKEYRANGEREQUEST.fields_by_name['charset'].message_type = _CHARSET
_STREAMKEYRANGERESPONSE.fields_by_name['binlog_transaction'].message_type = _BINLOGTRANSACTION
_STREAMTABLESREQUEST.fields_by_name['charset'].message_type = _CHARSET
_STREAMTABLESRESPONSE.fields_by_name['binlog_transaction'].message_type = _BINLOGTRANSACTION
_BINLOGFILTER.fields_by_name['key_range'].message_type = topodata__pb2._KEYRANGE
DESCRIPTOR.message_types_by_name['Charset'] = _CHARSET
DESCRIPTOR.message_types_by_name['BinlogTransaction'] = _BINLOGTRANSACTION
DESCRIPTOR.message_types_by_name['StreamKeyRangeRequest'] = _STREAMKEYRANGEREQUEST
DESCRIPTOR.message_types_by_name['StreamKeyRangeResponse'] = _STREAMKEYRANGERESPONSE
DESCRIPTOR.message_types_by_name['StreamTablesRequest'] = _STREAMTABLESREQUEST
DESCRIPTOR.message_types_by_name['StreamTablesResponse'] = _STREAMTABLESRESPONSE
DESCRIPTOR.message_types_by_name['BinlogFilter'] = _BINLOGFILTER
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

Charset = _reflection.GeneratedProtocolMessageType('Charset', (_message.Message,), dict(
  DESCRIPTOR = _CHARSET,
  __module__ = 'binlogdata_pb2'
  # @@protoc_insertion_point(class_scope:binlogdata.Charset)
  ))
_sym_db.RegisterMessage(Charset)

BinlogTransaction = _reflection.GeneratedProtocolMessageType('BinlogTransaction', (_message.Message,), dict(

  Statement = _reflection.GeneratedProtocolMessageType('Statement', (_message.Message,), dict(
    DESCRIPTOR = _BINLOGTRANSACTION_STATEMENT,
    __module__ = 'binlogdata_pb2'
    # @@protoc_insertion_point(class_scope:binlogdata.BinlogTransaction.Statement)
    ))
  ,
  DESCRIPTOR = _BINLOGTRANSACTION,
  __module__ = 'binlogdata_pb2'
  # @@protoc_insertion_point(class_scope:binlogdata.BinlogTransaction)
  ))
_sym_db.RegisterMessage(BinlogTransaction)
_sym_db.RegisterMessage(BinlogTransaction.Statement)

StreamKeyRangeRequest = _reflection.GeneratedProtocolMessageType('StreamKeyRangeRequest', (_message.Message,), dict(
  DESCRIPTOR = _STREAMKEYRANGEREQUEST,
  __module__ = 'binlogdata_pb2'
  # @@protoc_insertion_point(class_scope:binlogdata.StreamKeyRangeRequest)
  ))
_sym_db.RegisterMessage(StreamKeyRangeRequest)

StreamKeyRangeResponse = _reflection.GeneratedProtocolMessageType('StreamKeyRangeResponse', (_message.Message,), dict(
  DESCRIPTOR = _STREAMKEYRANGERESPONSE,
  __module__ = 'binlogdata_pb2'
  # @@protoc_insertion_point(class_scope:binlogdata.StreamKeyRangeResponse)
  ))
_sym_db.RegisterMessage(StreamKeyRangeResponse)

StreamTablesRequest = _reflection.GeneratedProtocolMessageType('StreamTablesRequest', (_message.Message,), dict(
  DESCRIPTOR = _STREAMTABLESREQUEST,
  __module__ = 'binlogdata_pb2'
  # @@protoc_insertion_point(class_scope:binlogdata.StreamTablesRequest)
  ))
_sym_db.RegisterMessage(StreamTablesRequest)

StreamTablesResponse = _reflection.GeneratedProtocolMessageType('StreamTablesResponse', (_message.Message,), dict(
  DESCRIPTOR = _STREAMTABLESRESPONSE,
  __module__ = 'binlogdata_pb2'
  # @@protoc_insertion_point(class_scope:binlogdata.StreamTablesResponse)
  ))
_sym_db.RegisterMessage(StreamTablesResponse)

BinlogFilter = _reflection.GeneratedProtocolMessageType('BinlogFilter', (_message.Message,), dict(
  DESCRIPTOR = _BINLOGFILTER,
  __module__ = 'binlogdata_pb2'
  # @@protoc_insertion_point(class_scope:binlogdata.BinlogFilter)
  ))
_sym_db.RegisterMessage(BinlogFilter)


DESCRIPTOR.has_options = True
DESCRIPTOR._options = _descriptor._ParseOptions(descriptor_pb2.FileOptions(), _b('Z\'vitess.io/vitess/go/vt/proto/binlogdata'))
# @@protoc_insertion_point(module_scope)
