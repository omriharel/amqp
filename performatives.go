package amqp

import "errors"

// Preformative Types
const (
	PerformativeEmpty       = 0xff // used to indicate empty payload, not part of spec
	PerformativeOpen        = 0x10
	PreformativeBegin       = 0x11
	PreformativeAttach      = 0x12
	PreformativeFlow        = 0x13
	PreformativeTransfer    = 0x14
	PreformativeDisposition = 0x15
	PreformativeDetach      = 0x16
	PreformativeEnd         = 0x17
	PreformativeClose       = 0x18

	TypeSource = 0x28
	TypeTarget = 0x29
	TypeError  = 0x1d
)

func preformativeType(payload []byte) (uint8, error) {
	if len(payload) == 0 {
		return PerformativeEmpty, nil
	}

	if len(payload) < 3 || payload[0] != 0 || payload[1] != Smallulong {
		return 0, errors.New("invalid preformative header")
	}

	return payload[2], nil
}

/*
<type name="open" class="composite" source="list" provides="frame">
    <descriptor name="amqp:open:list" code="0x00000000:0x00000010"/>
    <field name="container-id" type="string" mandatory="true"/>
    <field name="hostname" type="string"/>
    <field name="max-frame-size" type="uint" default="4294967295"/>
    <field name="channel-max" type="ushort" default="65535"/>
    <field name="idle-time-out" type="milliseconds"/>
    <field name="outgoing-locales" type="ietf-language-tag" multiple="true"/>
    <field name="incoming-locales" type="ietf-language-tag" multiple="true"/>
    <field name="offered-capabilities" type="symbol" multiple="true"/>
    <field name="desired-capabilities" type="symbol" multiple="true"/>
    <field name="properties" type="fields"/>
</type>
*/
type Open struct {
	ContainerID         string // required
	Hostname            string
	MaxFrameSize        uint32       // default: 4294967295
	ChannelMax          uint16       // default: 65535
	IdleTimeout         Milliseconds // from milliseconds
	OutgoingLocales     []Symbol
	IncomingLocales     []Symbol
	OfferedCapabilities []Symbol
	DesiredCapabilities []Symbol
	Properties          map[string]string // TODO: implement marshal/unmarshal
}

func (o *Open) MarshalBinary() ([]byte, error) {
	fields := []field{
		{value: o.ContainerID, omit: false},
		{value: o.Hostname, omit: o.Hostname == ""},
		{value: o.MaxFrameSize, omit: o.MaxFrameSize == 0},
		{value: o.ChannelMax, omit: o.ChannelMax == 0},
		{value: o.IdleTimeout, omit: o.IdleTimeout.Duration == 0},
		{value: o.OutgoingLocales, omit: len(o.OutgoingLocales) == 0},
		{value: o.IncomingLocales, omit: len(o.IncomingLocales) == 0},
		{value: o.OfferedCapabilities, omit: len(o.OfferedCapabilities) == 0},
		{value: o.DesiredCapabilities, omit: len(o.DesiredCapabilities) == 0},
		{value: o.Properties, omit: len(o.Properties) == 0},
	}
	return marshalComposite(PerformativeOpen, fields...)
}

func (o *Open) UnmarshalBinary(r byteReader) error {
	return unmarshalComposite(r, PerformativeOpen,
		&o.ContainerID,
		&o.Hostname,
		&o.MaxFrameSize,
		&o.ChannelMax,
		&o.IdleTimeout,
		&o.OutgoingLocales,
		&o.IncomingLocales,
		&o.OfferedCapabilities,
		&o.DesiredCapabilities,
	)
}

/*
<type name="begin" class="composite" source="list" provides="frame">
    <descriptor name="amqp:begin:list" code="0x00000000:0x00000011"/>
    <field name="remote-channel" type="ushort"/>
    <field name="next-outgoing-id" type="transfer-number" mandatory="true"/>
    <field name="incoming-window" type="uint" mandatory="true"/>
    <field name="outgoing-window" type="uint" mandatory="true"/>
    <field name="handle-max" type="handle" default="4294967295"/>
    <field name="offered-capabilities" type="symbol" multiple="true"/>
    <field name="desired-capabilities" type="symbol" multiple="true"/>
    <field name="properties" type="fields"/>
</type>
*/
type Begin struct {
	// the remote channel for this session
	// If a session is locally initiated, the remote-channel MUST NOT be set.
	// When an endpoint responds to a remotely initiated session, the remote-channel
	// MUST be set to the channel on which the remote session sent the begin.
	RemoteChannel uint16

	// the transfer-id of the first transfer id the sender will send
	NextOutgoingID uint32 // required, sequence number http://www.ietf.org/rfc/rfc1982.txt

	// the initial incoming-window of the sender
	IncomingWindow uint32 // required

	// the initial outgoing-window of the sender
	OutgoingWindow uint32 // required

	// the maximum handle value that can be used on the session
	// The handle-max value is the highest handle value that can be
	// used on the session. A peer MUST NOT attempt to attach a link
	// using a handle value outside the range that its partner can handle.
	// A peer that receives a handle outside the supported range MUST
	// close the connection with the framing-error error-code.
	HandleMax uint32 // default 4294967295

	// the extension capabilities the sender supports
	// http://www.amqp.org/specification/1.0/session-capabilities
	OfferedCapabilities []Symbol

	// the extension capabilities the sender can use if the receiver supports them
	// The sender MUST NOT attempt to use any capability other than those it
	// has declared in desired-capabilities field.
	DesiredCapabilities []Symbol

	// session properties
	// http://www.amqp.org/specification/1.0/session-properties
	Properties map[string]interface{}
}

func (b *Begin) MarshalBinary() ([]byte, error) {
	fields := []field{
		{value: b.RemoteChannel, omit: b.RemoteChannel == 0},
		{value: b.NextOutgoingID, omit: false},
		{value: b.IncomingWindow, omit: false},
		{value: b.OutgoingWindow, omit: false},
		{value: b.HandleMax, omit: b.HandleMax == 0},
		{value: b.OfferedCapabilities, omit: len(b.OfferedCapabilities) == 0},
		{value: b.DesiredCapabilities, omit: len(b.DesiredCapabilities) == 0},
	}
	return marshalComposite(PreformativeBegin, fields...)
}

func (b *Begin) UnmarshalBinary(r byteReader) error {
	return unmarshalComposite(r, PreformativeBegin,
		&b.RemoteChannel,
		&b.NextOutgoingID,
		&b.IncomingWindow,
		&b.OutgoingWindow,
		&b.HandleMax,
		&b.OfferedCapabilities,
		&b.DesiredCapabilities,
	)
}

/*
<type name="attach" class="composite" source="list" provides="frame">
    <descriptor name="amqp:attach:list" code="0x00000000:0x00000012"/>
    <field name="name" type="string" mandatory="true"/>
    <field name="handle" type="handle" mandatory="true"/>
    <field name="role" type="role" mandatory="true"/>
    <field name="snd-settle-mode" type="sender-settle-mode" default="mixed"/>
    <field name="rcv-settle-mode" type="receiver-settle-mode" default="first"/>
    <field name="source" type="*" requires="source"/>
    <field name="target" type="*" requires="target"/>
    <field name="unsettled" type="map"/>
    <field name="incomplete-unsettled" type="boolean" default="false"/>
    <field name="initial-delivery-count" type="sequence-no"/>
    <field name="max-message-size" type="ulong"/>
    <field name="offered-capabilities" type="symbol" multiple="true"/>
    <field name="desired-capabilities" type="symbol" multiple="true"/>
    <field name="properties" type="fields"/>
</type>
*/
type Attach struct {
	// the name of the link
	//
	// This name uniquely identifies the link from the container of the source
	//to the container of the target node, e.g., if the container of the source
	// node is A, and the container of the target node is B, the link MAY be
	// globally identified by the (ordered) tuple (A,B,<name>).
	Name string // required

	// the handle for the link while attached
	//
	// The numeric handle assigned by the the peer as a shorthand to refer to the
	// link in all performatives that reference the link until the it is detached.
	//
	// The handle MUST NOT be used for other open links. An attempt to attach using
	// a handle which is already associated with a link MUST be responded to with
	// an immediate close carrying a handle-in-use session-error.
	//
	// To make it easier to monitor AMQP link attach frames, it is RECOMMENDED that
	// implementations always assign the lowest available handle to this field.
	//
	// The two endpoints MAY potentially use different handles to refer to the same link.
	// Link handles MAY be reused once a link is closed for both send and receive.
	Handle uint32 // required

	// role of the link endpoint
	//
	// The role being played by the peer, i.e., whether the peer is the sender or the
	// receiver of messages on the link.
	Role bool // required, true=reciever / false=sender

	// settlement policy for the sender
	//
	// The delivery settlement policy for the sender. When set at the receiver this
	// indicates the desired value for the settlement mode at the sender. When set
	// at the sender this indicates the actual settlement mode in use. The sender
	// SHOULD respect the receiver's desired settlement mode if the receiver initiates
	// the attach exchange and the sender supports the desired mode.
	//
	// 0: unsettled - The sender will send all deliveries initially unsettled to the receiver.
	// 1: settled - The sender will send all deliveries settled to the receiver.
	// 2: mixed - The sender MAY send a mixture of settled and unsettled deliveries to the receiver.
	SenderSettleMode uint8

	// the settlement policy of the receiver
	//
	// The delivery settlement policy for the receiver. When set at the sender this
	// indicates the desired value for the settlement mode at the receiver.
	// When set at the receiver this indicates the actual settlement mode in use.
	// The receiver SHOULD respect the sender's desired settlement mode if the sender
	// initiates the attach exchange and the receiver supports the desired mode.
	//
	// 0: first - The receiver will spontaneously settle all incoming transfers.
	// 1: second - The receiver will only settle after sending the disposition to
	//             the sender and receiving a disposition indicating settlement of
	//             the delivery from the sender.
	ReceiverSettleMode uint8

	// the source for messages
	//
	// If no source is specified on an outgoing link, then there is no source currently
	// attached to the link. A link with no source will never produce outgoing messages.
	Source *Source

	// the target for messages
	//
	// If no target is specified on an incoming link, then there is no target currently
	// attached to the link. A link with no target will never permit incoming messages.
	Target *Target

	// unsettled delivery state
	//
	// This is used to indicate any unsettled delivery states when a suspended link is
	// resumed. The map is keyed by delivery-tag with values indicating the delivery state.
	// The local and remote delivery states for a given delivery-tag MUST be compared to
	// resolve any in-doubt deliveries. If necessary, deliveries MAY be resent, or resumed
	// based on the outcome of this comparison. See subsection 2.6.13.
	//
	// If the local unsettled map is too large to be encoded within a frame of the agreed
	// maximum frame size then the session MAY be ended with the frame-size-too-small error.
	// The endpoint SHOULD make use of the ability to send an incomplete unsettled map
	// (see below) to avoid sending an error.
	//
	// The unsettled map MUST NOT contain null valued keys.
	//
	// When reattaching (as opposed to resuming), the unsettled map MUST be null.
	Unsettled map[string]uint

	// If set to true this field indicates that the unsettled map provided is not complete.
	// When the map is incomplete the recipient of the map cannot take the absence of a
	// delivery tag from the map as evidence of settlement. On receipt of an incomplete
	// unsettled map a sending endpoint MUST NOT send any new deliveries (i.e. deliveries
	// where resume is not set to true) to its partner (and a receiving endpoint which sent
	// an incomplete unsettled map MUST detach with an error on receiving a transfer which
	// does not have the resume flag set to true).
	//
	// Note that if this flag is set to true then the endpoints MUST detach and reattach at
	// least once in order to send new deliveries. This flag can be useful when there are
	// too many entries in the unsettled map to fit within a single frame. An endpoint can
	// attach, resume, settle, and detach until enough unsettled state has been cleared for
	// an attach where this flag is set to false.
	IncompleteUnsettled bool // default: false

	// the sender's initial value for delivery-count
	//
	// This MUST NOT be null if role is sender, and it is ignored if the role is receiver.
	InitialDeliveryCount uint32 // sequence number

	// the maximum message size supported by the link endpoint
	//
	// This field indicates the maximum message size supported by the link endpoint.
	// Any attempt to deliver a message larger than this results in a message-size-exceeded
	// link-error. If this field is zero or unset, there is no maximum size imposed by the
	// link endpoint.
	MaxMessageSize uint64

	// the extension capabilities the sender supports
	// http://www.amqp.org/specification/1.0/link-capabilities
	OfferedCapabilities []Symbol

	// the extension capabilities the sender can use if the receiver supports them
	//
	// The sender MUST NOT attempt to use any capability other than those it
	// has declared in desired-capabilities field.
	DesiredCapabilities []Symbol

	// link properties
	// http://www.amqp.org/specification/1.0/link-properties
	Properties map[string]interface{}
}

func (a *Attach) MarshalBinary() ([]byte, error) {
	fields := []field{
		{value: a.Name, omit: false},
		{value: a.Handle, omit: false},
		{value: a.Role, omit: false},
		{value: a.SenderSettleMode, omit: a.SenderSettleMode == 2},
		{value: a.ReceiverSettleMode, omit: a.ReceiverSettleMode == 0},
		{value: a.Source, omit: a.Source == nil},
		{value: a.Target, omit: a.Target == nil},
		{value: a.Unsettled, omit: len(a.Unsettled) == 0},
		{value: a.IncompleteUnsettled, omit: !a.IncompleteUnsettled},
		{value: a.InitialDeliveryCount, omit: a.InitialDeliveryCount == 0},
		{value: a.MaxMessageSize, omit: a.MaxMessageSize == 0},
		{value: a.OfferedCapabilities, omit: len(a.OfferedCapabilities) == 0},
		{value: a.DesiredCapabilities, omit: len(a.DesiredCapabilities) == 0},
		{value: a.Properties, omit: len(a.Properties) == 0},
	}
	return marshalComposite(PreformativeAttach, fields...)
}

func (a *Attach) UnmarshalBinary(r byteReader) error {
	return unmarshalComposite(r, PreformativeAttach,
		&a.Name,
		&a.Handle,
		&a.Role,
		&a.SenderSettleMode,
		&a.ReceiverSettleMode,
		&a.Source,
		&a.Target,
		&a.Unsettled,
		&a.IncompleteUnsettled,
		&a.InitialDeliveryCount,
		&a.MaxMessageSize,
		&a.OfferedCapabilities,
		&a.DesiredCapabilities,
		&a.Properties,
	)
}

/*
<type name="source" class="composite" source="list" provides="source">
    <descriptor name="amqp:source:list" code="0x00000000:0x00000028"/>
    <field name="address" type="*" requires="address"/>
    <field name="durable" type="terminus-durability" default="none"/>
    <field name="expiry-policy" type="terminus-expiry-policy" default="session-end"/>
    <field name="timeout" type="seconds" default="0"/>
    <field name="dynamic" type="boolean" default="false"/>
    <field name="dynamic-node-properties" type="node-properties"/>
    <field name="distribution-mode" type="symbol" requires="distribution-mode"/>
    <field name="filter" type="filter-set"/>
    <field name="default-outcome" type="*" requires="outcome"/>
    <field name="outcomes" type="symbol" multiple="true"/>
    <field name="capabilities" type="symbol" multiple="true"/>
</type>
*/
type Source struct {
	// the address of the source
	//
	// The address of the source MUST NOT be set when sent on a attach frame sent by
	// the receiving link endpoint where the dynamic flag is set to true (that is where
	// the receiver is requesting the sender to create an addressable node).
	//
	// The address of the source MUST be set when sent on a attach frame sent by the
	// sending link endpoint where the dynamic flag is set to true (that is where the
	// sender has created an addressable node at the request of the receiver and is now
	// communicating the address of that created node). The generated name of the address
	// SHOULD include the link name and the container-id of the remote container to allow
	// for ease of identification.
	Address string

	// indicates the durability of the terminus
	//
	// Indicates what state of the terminus will be retained durably: the state of durable
	// messages, only existence and configuration of the terminus, or no state at all.
	//
	// 0: none
	// 1: configuration
	// 2: unsettled-state
	Durable uint32

	// the expiry policy of the source
	//
	// link-detach: The expiry timer starts when terminus is detached.
	// session-end: The expiry timer starts when the most recently associated session is
	//              ended.
	// connection-close: The expiry timer starts when most recently associated connection
	//                   is closed.
	// never: The terminus never expires.
	ExpiryPolicy Symbol

	// duration that an expiring source will be retained
	//
	// The source starts expiring as indicated by the expiry-policy.
	Timeout uint32 // seconds

	// request dynamic creation of a remote node
	//
	// When set to true by the receiving link endpoint, this field constitutes a request
	// for the sending peer to dynamically create a node at the source. In this case the
	// address field MUST NOT be set.
	//
	// When set to true by the sending link endpoint this field indicates creation of a
	// dynamically created node. In this case the address field will contain the address
	// of the created node. The generated address SHOULD include the link name and other
	// available information on the initiator of the request (such as the remote
	// container-id) in some recognizable form for ease of traceability.
	Dynamic bool

	// properties of the dynamically created node
	//
	// If the dynamic field is not set to true this field MUST be left unset.
	//
	// When set by the receiving link endpoint, this field contains the desired
	// properties of the node the receiver wishes to be created. When set by the
	// sending link endpoint this field contains the actual properties of the
	// dynamically created node. See subsection 3.5.9 for standard node properties.
	// http://www.amqp.org/specification/1.0/node-properties
	//
	// lifetime-policy: The lifetime of a dynamically generated node.
	//					Definitionally, the lifetime will never be less than the lifetime
	//					of the link which caused its creation, however it is possible to
	//					extend the lifetime of dynamically created node using a lifetime
	//					policy. The value of this entry MUST be of a type which provides
	//					the lifetime-policy archetype. The following standard
	//					lifetime-policies are defined below: delete-on-close,
	//					delete-on-no-links, delete-on-no-messages or
	//					delete-on-no-links-or-messages.
	// supported-dist-modes: The distribution modes that the node supports.
	//					The value of this entry MUST be one or more symbols which are valid
	//					distribution-modes. That is, the value MUST be of the same type as
	//					would be valid in a field defined with the following attributes:
	//						type="symbol" multiple="true" requires="distribution-mode"
	DynamicNodeProperties map[Symbol]interface{}

	// the distribution mode of the link
	//
	// This field MUST be set by the sending end of the link if the endpoint supports more
	// than one distribution-mode. This field MAY be set by the receiving end of the link
	// to indicate a preference when a node supports multiple distribution modes.
	DistributionMode Symbol

	// a set of predicates to filter the messages admitted onto the link
	//
	// The receiving endpoint sets its desired filter, the sending endpoint sets the filter
	// actually in place (including any filters defaulted at the node). The receiving
	// endpoint MUST check that the filter in place meets its needs and take responsibility
	// for detaching if it does not.
	Filter map[Symbol]interface{}

	// default outcome for unsettled transfers
	//
	// Indicates the outcome to be used for transfers that have not reached a terminal
	// state at the receiver when the transfer is settled, including when the source
	// is destroyed. The value MUST be a valid outcome (e.g., released or rejected).
	DefaultOutcome interface{}

	// descriptors for the outcomes that can be chosen on this link
	//
	// The values in this field are the symbolic descriptors of the outcomes that can
	// be chosen on this link. This field MAY be empty, indicating that the default-outcome
	// will be assumed for all message transfers (if the default-outcome is not set, and no
	// outcomes are provided, then the accepted outcome MUST be supported by the source).
	//
	// When present, the values MUST be a symbolic descriptor of a valid outcome,
	// e.g., "amqp:accepted:list".
	Outcomes []Symbol

	// the extension capabilities the sender supports/desires
	//
	// http://www.amqp.org/specification/1.0/source-capabilities
	Capabilities []Symbol
}

func (s *Source) MarshalBinary() ([]byte, error) {
	fields := []field{
		{value: s.Address, omit: s.Address == ""},
		{value: s.Durable, omit: s.Durable == 0},
		{value: s.ExpiryPolicy, omit: s.ExpiryPolicy == ""},
		{value: s.Timeout, omit: s.Timeout == 0},
		{value: s.Dynamic, omit: !s.Dynamic},
		{value: s.DynamicNodeProperties, omit: len(s.DynamicNodeProperties) == 0},
		{value: s.DistributionMode, omit: s.DistributionMode == ""},
		{value: s.Filter, omit: len(s.Filter) == 0},
		{value: s.DefaultOutcome, omit: s.DefaultOutcome == nil},
		{value: s.Outcomes, omit: len(s.Outcomes) == 0},
		{value: s.Capabilities, omit: len(s.Capabilities) == 0},
	}
	return marshalComposite(TypeSource, fields...)
}

func (s *Source) UnmarshalBinary(r byteReader) error {
	return unmarshalComposite(r, TypeSource,
		&s.Address,
		&s.Durable,
		&s.ExpiryPolicy,
		&s.Timeout,
		&s.Dynamic,
		&s.DynamicNodeProperties,
		&s.DistributionMode,
		&s.Filter,
		&s.DefaultOutcome,
		&s.Outcomes,
		&s.Capabilities,
	)
}

/*
<type name="target" class="composite" source="list" provides="target">
    <descriptor name="amqp:target:list" code="0x00000000:0x00000029"/>
    <field name="address" type="*" requires="address"/>
    <field name="durable" type="terminus-durability" default="none"/>
    <field name="expiry-policy" type="terminus-expiry-policy" default="session-end"/>
    <field name="timeout" type="seconds" default="0"/>
    <field name="dynamic" type="boolean" default="false"/>
    <field name="dynamic-node-properties" type="node-properties"/>
    <field name="capabilities" type="symbol" multiple="true"/>
</type>
*/
type Target struct {
	// the address of the target
	//
	// The address of the target MUST NOT be set when sent on a attach frame sent by
	// the sending link endpoint where the dynamic flag is set to true (that is where
	// the sender is requesting the receiver to create an addressable node).
	//
	// The address of the source MUST be set when sent on a attach frame sent by the
	// receiving link endpoint where the dynamic flag is set to true (that is where
	// the receiver has created an addressable node at the request of the sender and
	// is now communicating the address of that created node). The generated name of
	// the address SHOULD include the link name and the container-id of the remote
	// container to allow for ease of identification.
	Address string

	// indicates the durability of the terminus
	//
	// Indicates what state of the terminus will be retained durably: the state of durable
	// messages, only existence and configuration of the terminus, or no state at all.
	//
	// 0: none
	// 1: configuration
	// 2: unsettled-state
	Durable uint32

	// the expiry policy of the target
	//
	// link-detach: The expiry timer starts when terminus is detached.
	// session-end: The expiry timer starts when the most recently associated session is
	//              ended.
	// connection-close: The expiry timer starts when most recently associated connection
	//                   is closed.
	// never: The terminus never expires.
	ExpiryPolicy Symbol

	// duration that an expiring target will be retained
	//
	// The target starts expiring as indicated by the expiry-policy.
	Timeout uint32 // seconds

	// request dynamic creation of a remote node
	//
	// When set to true by the sending link endpoint, this field constitutes a request
	// for the receiving peer to dynamically create a node at the target. In this case
	// the address field MUST NOT be set.
	//
	// When set to true by the receiving link endpoint this field indicates creation of
	// a dynamically created node. In this case the address field will contain the
	// address of the created node. The generated address SHOULD include the link name
	// and other available information on the initiator of the request (such as the
	// remote container-id) in some recognizable form for ease of traceability.
	Dynamic bool

	// properties of the dynamically created node
	//
	// If the dynamic field is not set to true this field MUST be left unset.
	//
	// When set by the sending link endpoint, this field contains the desired
	// properties of the node the sender wishes to be created. When set by the
	// receiving link endpoint this field contains the actual properties of the
	// dynamically created node. See subsection 3.5.9 for standard node properties.
	// http://www.amqp.org/specification/1.0/node-properties
	//
	// lifetime-policy: The lifetime of a dynamically generated node.
	//					Definitionally, the lifetime will never be less than the lifetime
	//					of the link which caused its creation, however it is possible to
	//					extend the lifetime of dynamically created node using a lifetime
	//					policy. The value of this entry MUST be of a type which provides
	//					the lifetime-policy archetype. The following standard
	//					lifetime-policies are defined below: delete-on-close,
	//					delete-on-no-links, delete-on-no-messages or
	//					delete-on-no-links-or-messages.
	// supported-dist-modes: The distribution modes that the node supports.
	//					The value of this entry MUST be one or more symbols which are valid
	//					distribution-modes. That is, the value MUST be of the same type as
	//					would be valid in a field defined with the following attributes:
	//						type="symbol" multiple="true" requires="distribution-mode"
	DynamicNodeProperties map[Symbol]interface{}

	// the extension capabilities the sender supports/desires
	//
	// http://www.amqp.org/specification/1.0/target-capabilities
	Capabilities []Symbol
}

func (t *Target) MarshalBinary() ([]byte, error) {
	fields := []field{
		{value: t.Address, omit: t.Address == ""},
		{value: t.Durable, omit: t.Durable == 0},
		{value: t.ExpiryPolicy, omit: t.ExpiryPolicy == ""},
		{value: t.Timeout, omit: t.Timeout == 0},
		{value: t.Dynamic, omit: !t.Dynamic},
		{value: t.DynamicNodeProperties, omit: len(t.DynamicNodeProperties) == 0},
		{value: t.Capabilities, omit: len(t.Capabilities) == 0},
	}
	return marshalComposite(TypeSource, fields...)
}

func (t *Target) UnmarshalBinary(r byteReader) error {
	return unmarshalComposite(r, TypeTarget,
		&t.Address,
		&t.Durable,
		&t.ExpiryPolicy,
		&t.Timeout,
		&t.Dynamic,
		&t.DynamicNodeProperties,
		&t.Capabilities,
	)
}

/*
<type name="flow" class="composite" source="list" provides="frame">
    <descriptor name="amqp:flow:list" code="0x00000000:0x00000013"/>
    <field name="next-incoming-id" type="transfer-number"/>
    <field name="incoming-window" type="uint" mandatory="true"/>
    <field name="next-outgoing-id" type="transfer-number" mandatory="true"/>
    <field name="outgoing-window" type="uint" mandatory="true"/>
    <field name="handle" type="handle"/>
    <field name="delivery-count" type="sequence-no"/>
    <field name="link-credit" type="uint"/>
    <field name="available" type="uint"/>
    <field name="drain" type="boolean" default="false"/>
    <field name="echo" type="boolean" default="false"/>
    <field name="properties" type="fields"/>
</type>
*/
type Flow struct {
	// Identifies the expected transfer-id of the next incoming transfer frame.
	// This value MUST be set if the peer has received the begin frame for the
	// session, and MUST NOT be set if it has not. See subsection 2.5.6 for more details.
	NextIncomingID *uint32 // sequence number

	// Defines the maximum number of incoming transfer frames that the endpoint
	// can currently receive. See subsection 2.5.6 for more details.
	IncomingWindow uint32 // required

	// The transfer-id that will be assigned to the next outgoing transfer frame.
	// See subsection 2.5.6 for more details.
	NextOutgoingID uint32 // sequence number

	// Defines the maximum number of outgoing transfer frames that the endpoint
	// could potentially currently send, if it was not constrained by restrictions
	// imposed by its peer's incoming-window. See subsection 2.5.6 for more details.
	OutgoingWindow uint32

	// If set, indicates that the flow frame carries flow state information for the local
	// link endpoint associated with the given handle. If not set, the flow frame is
	// carrying only information pertaining to the session endpoint.
	//
	// If set to a handle that is not currently associated with an attached link,
	// the recipient MUST respond by ending the session with an unattached-handle
	// session error.
	Handle *uint32

	// The delivery-count is initialized by the sender when a link endpoint is created,
	// and is incremented whenever a message is sent. Only the sender MAY independently
	// modify this field. The receiver's value is calculated based on the last known
	// value from the sender and any subsequent messages received on the link. Note that,
	// despite its name, the delivery-count is not a count but a sequence number
	// initialized at an arbitrary point by the sender.
	//
	// When the handle field is not set, this field MUST NOT be set.
	//
	// When the handle identifies that the flow state is being sent from the sender link
	// endpoint to receiver link endpoint this field MUST be set to the current
	// delivery-count of the link endpoint.
	//
	// When the flow state is being sent from the receiver endpoint to the sender endpoint
	// this field MUST be set to the last known value of the corresponding sending endpoint.
	// In the event that the receiving link endpoint has not yet seen the initial attach
	// frame from the sender this field MUST NOT be set.
	DeliveryCount *uint32 // sequence number

	// the current maximum number of messages that can be received
	//
	// The current maximum number of messages that can be handled at the receiver endpoint
	// of the link. Only the receiver endpoint can independently set this value. The sender
	// endpoint sets this to the last known value seen from the receiver.
	// See subsection 2.6.7 for more details.
	//
	// When the handle field is not set, this field MUST NOT be set.
	LinkCredit *uint32

	// the number of available messages
	//
	// The number of messages awaiting credit at the link sender endpoint. Only the sender
	// can independently set this value. The receiver sets this to the last known value seen
	// from the sender. See subsection 2.6.7 for more details.
	//
	// When the handle field is not set, this field MUST NOT be set.
	Available *uint32

	// indicates drain mode
	//
	// When flow state is sent from the sender to the receiver, this field contains the
	// actual drain mode of the sender. When flow state is sent from the receiver to the
	// sender, this field contains the desired drain mode of the receiver.
	// See subsection 2.6.7 for more details.
	//
	// When the handle field is not set, this field MUST NOT be set.
	Drain bool

	// request state from partner
	//
	// If set to true then the receiver SHOULD send its state at the earliest convenient
	// opportunity.
	//
	// If set to true, and the handle field is not set, then the sender only requires
	// session endpoint state to be echoed, however, the receiver MAY fulfil this requirement
	// by sending a flow performative carrying link-specific state (since any such flow also
	// carries session state).
	//
	// If a sender makes multiple requests for the same state before the receiver can reply,
	// the receiver MAY send only one flow in return.
	//
	// Note that if a peer responds to echo requests with flows which themselves have the
	// echo field set to true, an infinite loop could result if its partner adopts the same
	// policy (therefore such a policy SHOULD be avoided).
	Echo bool

	// link state properties
	// http://www.amqp.org/specification/1.0/link-state-properties
	Properties map[string]interface{}
}

func (f *Flow) MarshalBinary() ([]byte, error) {
	fields := []field{
		{value: f.NextIncomingID, omit: f.NextIncomingID == nil},
		{value: f.IncomingWindow, omit: false},
		{value: f.NextOutgoingID, omit: false},
		{value: f.OutgoingWindow, omit: false},
		{value: f.Handle, omit: f.Handle == nil},
		{value: f.DeliveryCount, omit: f.DeliveryCount == nil},
		{value: f.LinkCredit, omit: f.LinkCredit == nil},
		{value: f.Available, omit: f.Available == nil},
		{value: f.Drain, omit: !f.Drain},
		{value: f.Echo, omit: !f.Echo},
		{value: f.Properties, omit: len(f.Properties) == 0},
	}
	return marshalComposite(PreformativeFlow, fields...)
}

func (f *Flow) UnmarshalBinary(r byteReader) error {
	return unmarshalComposite(r, PreformativeFlow,
		f.NextIncomingID,
		&f.IncomingWindow,
		&f.NextOutgoingID,
		&f.OutgoingWindow,
		f.Handle,
		f.DeliveryCount,
		f.LinkCredit,
		f.Available,
		&f.Drain,
		&f.Echo,
		&f.Properties,
	)
}

/*
<type name="transfer" class="composite" source="list" provides="frame">
    <descriptor name="amqp:transfer:list" code="0x00000000:0x00000014"/>
    <field name="handle" type="handle" mandatory="true"/>
    <field name="delivery-id" type="delivery-number"/>
    <field name="delivery-tag" type="delivery-tag"/>
    <field name="message-format" type="message-format"/>
    <field name="settled" type="boolean"/>
    <field name="more" type="boolean" default="false"/>
    <field name="rcv-settle-mode" type="receiver-settle-mode"/>
    <field name="state" type="*" requires="delivery-state"/>
    <field name="resume" type="boolean" default="false"/>
    <field name="aborted" type="boolean" default="false"/>
    <field name="batchable" type="boolean" default="false"/>
</type>
*/
type Transfer struct {
	// Specifies the link on which the message is transferred.
	Handle uint32 // required

	// The delivery-id MUST be supplied on the first transfer of a multi-transfer
	// delivery. On continuation transfers the delivery-id MAY be omitted. It is
	// an error if the delivery-id on a continuation transfer differs from the
	// delivery-id on the first transfer of a delivery.
	DeliveryID *uint32 // sequence number

	// Uniquely identifies the delivery attempt for a given message on this link.
	// This field MUST be specified for the first transfer of a multi-transfer
	// message and can only be omitted for continuation transfers. It is an error
	// if the delivery-tag on a continuation transfer differs from the delivery-tag
	// on the first transfer of a delivery.
	DeliveryTag []byte // up to 32 bytes

	// This field MUST be specified for the first transfer of a multi-transfer message
	// and can only be omitted for continuation transfers. It is an error if the
	// message-format on a continuation transfer differs from the message-format on
	// the first transfer of a delivery.
	//
	// The upper three octets of a message format code identify a particular message
	// format. The lowest octet indicates the version of said message format. Any given
	// version of a format is forwards compatible with all higher versions.
	MessageFormat *uint32

	// If not set on the first (or only) transfer for a (multi-transfer) delivery,
	// then the settled flag MUST be interpreted as being false. For subsequent
	// transfers in a multi-transfer delivery if the settled flag is left unset then
	// it MUST be interpreted as true if and only if the value of the settled flag on
	// any of the preceding transfers was true; if no preceding transfer was sent with
	// settled being true then the value when unset MUST be taken as false.
	//
	// If the negotiated value for snd-settle-mode at attachment is settled, then this
	// field MUST be true on at least one transfer frame for a delivery (i.e., the
	// delivery MUST be settled at the sender at the point the delivery has been
	// completely transferred).
	//
	// If the negotiated value for snd-settle-mode at attachment is unsettled, then this
	// field MUST be false (or unset) on every transfer frame for a delivery (unless the
	// delivery is aborted).
	Settled bool

	// indicates that the message has more content
	//
	// Note that if both the more and aborted fields are set to true, the aborted flag
	// takes precedence. That is, a receiver SHOULD ignore the value of the more field
	// if the transfer is marked as aborted. A sender SHOULD NOT set the more flag to
	// true if it also sets the aborted flag to true.
	More bool

	// If first, this indicates that the receiver MUST settle the delivery once it has
	// arrived without waiting for the sender to settle first.
	//
	// If second, this indicates that the receiver MUST NOT settle until sending its
	// disposition to the sender and receiving a settled disposition from the sender.
	//
	// If not set, this value is defaulted to the value negotiated on link attach.
	//
	// If the negotiated link value is first, then it is illegal to set this field
	// to second.
	//
	// If the message is being sent settled by the sender, the value of this field
	// is ignored.
	//
	// The (implicit or explicit) value of this field does not form part of the
	// transfer state, and is not retained if a link is suspended and subsequently resumed.
	//
	// 0: first - The receiver will spontaneously settle all incoming transfers.
	// 1: second - The receiver will only settle after sending the disposition to
	//             the sender and receiving a disposition indicating settlement of
	//             the delivery from the sender.
	ReceiverSettleMode *uint8

	// the state of the delivery at the sender
	//
	// When set this informs the receiver of the state of the delivery at the sender.
	// This is particularly useful when transfers of unsettled deliveries are resumed
	// after resuming a link. Setting the state on the transfer can be thought of as
	// being equivalent to sending a disposition immediately before the transfer
	// performative, i.e., it is the state of the delivery (not the transfer) that
	// existed at the point the frame was sent.
	//
	// Note that if the transfer performative (or an earlier disposition performative
	// referring to the delivery) indicates that the delivery has attained a terminal
	// state, then no future transfer or disposition sent by the sender can alter that
	// terminal state.
	State interface{} // TODO: add delivery states

	// indicates a resumed delivery
	//
	// If true, the resume flag indicates that the transfer is being used to reassociate
	// an unsettled delivery from a dissociated link endpoint. See subsection 2.6.13
	// for more details.
	//
	// The receiver MUST ignore resumed deliveries that are not in its local unsettled map.
	// The sender MUST NOT send resumed transfers for deliveries not in its local
	// unsettled map.
	//
	// If a resumed delivery spans more than one transfer performative, then the resume
	// flag MUST be set to true on the first transfer of the resumed delivery. For
	// subsequent transfers for the same delivery the resume flag MAY be set to true,
	// or MAY be omitted.
	//
	// In the case where the exchange of unsettled maps makes clear that all message
	// data has been successfully transferred to the receiver, and that only the final
	// state (and potentially settlement) at the sender needs to be conveyed, then a
	// resumed delivery MAY carry no payload and instead act solely as a vehicle for
	// carrying the terminal state of the delivery at the sender.
	Resume bool

	// indicates that the message is aborted
	//
	// Aborted messages SHOULD be discarded by the recipient (any payload within the
	// frame carrying the performative MUST be ignored). An aborted message is
	// implicitly settled.
	Aborted bool

	// batchable hint
	//
	// If true, then the issuer is hinting that there is no need for the peer to urgently
	// communicate updated delivery state. This hint MAY be used to artificially increase
	// the amount of batching an implementation uses when communicating delivery states,
	// and thereby save bandwidth.
	//
	// If the message being delivered is too large to fit within a single frame, then the
	// setting of batchable to true on any of the transfer performatives for the delivery
	// is equivalent to setting batchable to true for all the transfer performatives for
	// the delivery.
	//
	// The batchable value does not form part of the transfer state, and is not retained
	// if a link is suspended and subsequently resumed.
	Batchable bool
}

func (t *Transfer) MarshalBinary() ([]byte, error) {
	fields := []field{
		{value: t.DeliveryID, omit: t.DeliveryID == nil},
		{value: t.DeliveryTag, omit: len(t.DeliveryTag) == 0},
		{value: t.MessageFormat, omit: t.MessageFormat == nil},
		{value: t.Settled, omit: !t.Settled},
		{value: t.More, omit: !t.More},
		{value: t.ReceiverSettleMode, omit: t.ReceiverSettleMode == nil},
		{value: t.State, omit: t.State == nil},
		{value: t.Resume, omit: !t.Resume},
		{value: t.Aborted, omit: !t.Aborted},
		{value: t.Batchable, omit: !t.Batchable},
	}
	return marshalComposite(PreformativeFlow, fields...)
}

func (t *Transfer) UnmarshalBinary(r byteReader) error {
	return unmarshalComposite(r, PreformativeTransfer,
		t.DeliveryID,
		&t.DeliveryTag,
		t.MessageFormat,
		&t.Settled,
		&t.More,
		t.ReceiverSettleMode,
		&t.State,
		&t.Resume,
		&t.Aborted,
		&t.Batchable,
	)
}

/*
<type name="disposition" class="composite" source="list" provides="frame">
    <descriptor name="amqp:disposition:list" code="0x00000000:0x00000015"/>
    <field name="role" type="role" mandatory="true"/>
    <field name="first" type="delivery-number" mandatory="true"/>
    <field name="last" type="delivery-number"/>
    <field name="settled" type="boolean" default="false"/>
    <field name="state" type="*" requires="delivery-state"/>
    <field name="batchable" type="boolean" default="false"/>
</type>
*/
type Disposition struct {
	// directionality of disposition
	//
	// The role identifies whether the disposition frame contains information about
	// sending link endpoints or receiving link endpoints.
	Role bool // required, true=reciever / false=sender

	// lower bound of deliveries
	//
	// Identifies the lower bound of delivery-ids for the deliveries in this set.
	First uint32 // required, sequence number

	// upper bound of deliveries
	//
	// Identifies the upper bound of delivery-ids for the deliveries in this set.
	// If not set, this is taken to be the same as first.
	Last *uint32 // sequence number

	// indicates deliveries are settled
	//
	// If true, indicates that the referenced deliveries are considered settled by
	// the issuing endpoint.
	Settled bool

	// indicates state of deliveries
	//
	// Communicates the state of all the deliveries referenced by this disposition.
	State interface{}

	// batchable hint
	//
	// If true, then the issuer is hinting that there is no need for the peer to
	// urgently communicate the impact of the updated delivery states. This hint
	// MAY be used to artificially increase the amount of batching an implementation
	// uses when communicating delivery states, and thereby save bandwidth.
	Batchable bool
}

func (d *Disposition) MarshalBinary() ([]byte, error) {
	fields := []field{
		{value: d.Role, omit: false},
		{value: d.First, omit: false},
		{value: d.Last, omit: d.Last == nil},
		{value: d.Settled, omit: !d.Settled},
		{value: d.State, omit: d.State == nil},
		{value: d.Batchable, omit: !d.Batchable},
	}
	return marshalComposite(PreformativeDisposition, fields...)
}

func (d *Disposition) UnmarshalBinary(r byteReader) error {
	return unmarshalComposite(r, PreformativeDisposition,
		&d.Role,
		&d.First,
		d.Last,
		&d.Settled,
		&d.State,
		&d.Batchable,
	)
}

/*
<type name="detach" class="composite" source="list" provides="frame">
    <descriptor name="amqp:detach:list" code="0x00000000:0x00000016"/>
    <field name="handle" type="handle" mandatory="true"/>
    <field name="closed" type="boolean" default="false"/>
    <field name="error" type="error"/>
</type>
*/
type Detach struct {
	// the local handle of the link to be detached
	Handle uint32 //required

	// if true then the sender has closed the link
	Closed bool

	// error causing the detach
	//
	// If set, this field indicates that the link is being detached due to an error
	// condition. The value of the field SHOULD contain details on the cause of the error.
	Error *Error
}

func (d *Detach) MarshalBinary() ([]byte, error) {
	fields := []field{
		{value: d.Handle, omit: false},
		{value: d.Closed, omit: !d.Closed},
		{value: d.Error, omit: d.Error == nil},
	}
	return marshalComposite(PreformativeDetach, fields...)
}

func (d *Detach) UnmarshalBinary(r byteReader) error {
	return unmarshalComposite(r, PreformativeDetach,
		&d.Handle,
		&d.Closed,
		d.Error,
	)
}

/*
<type name="error" class="composite" source="list">
    <descriptor name="amqp:error:list" code="0x00000000:0x0000001d"/>
    <field name="condition" type="symbol" requires="error-condition" mandatory="true"/>
    <field name="description" type="string"/>
    <field name="info" type="fields"/>
</type>
*/
type Error struct {
	// error condition
	//
	// A symbolic value indicating the error condition.
	Condition Symbol // required

	// descriptive text about the error condition
	//
	// This text supplies any supplementary details not indicated by the condition field.
	// This text can be logged as an aid to resolving issues.
	Description string

	// map carrying information about the error condition
	Info map[Symbol]interface{}
}

func (e *Error) MarshalBinary() ([]byte, error) {
	fields := []field{
		{value: e.Condition, omit: false},
		{value: e.Description, omit: e.Description == ""},
		{value: e.Info, omit: len(e.Info) == 0},
	}
	return marshalComposite(TypeError, fields...)
}

func (e *Error) UnmarshalBinary(r byteReader) error {
	return unmarshalComposite(r, TypeError,
		&e.Condition,
		&e.Description,
		&e.Info,
	)
}

/*
<type name="end" class="composite" source="list" provides="frame">
    <descriptor name="amqp:end:list" code="0x00000000:0x00000017"/>
    <field name="error" type="error"/>
</type>
*/
type End struct {
	// error causing the end
	//
	// If set, this field indicates that the session is being ended due to an error
	// condition. The value of the field SHOULD contain details on the cause of the error.
	Error *Error
}

func (e *End) MarshalBinary() ([]byte, error) {
	fields := []field{
		{value: e.Error, omit: e.Error == nil},
	}
	return marshalComposite(PreformativeEnd, fields...)
}

func (e *End) UnmarshalBinary(r byteReader) error {
	return unmarshalComposite(r, PreformativeEnd,
		e.Error,
	)
}

/*
<type name="close" class="composite" source="list" provides="frame">
    <descriptor name="amqp:close:list" code="0x00000000:0x00000018"/>
    <field name="error" type="error"/>
</type>
*/
type Close struct {
	// error causing the close
	//
	// If set, this field indicates that the session is being closed due to an error
	// condition. The value of the field SHOULD contain details on the cause of the error.
	Error *Error
}

func (c *Close) MarshalBinary() ([]byte, error) {
	fields := []field{
		{value: c.Error, omit: c.Error == nil},
	}
	return marshalComposite(PreformativeClose, fields...)
}

func (c *Close) UnmarshalBinary(r byteReader) error {
	return unmarshalComposite(r, PreformativeClose,
		c.Error,
	)
}