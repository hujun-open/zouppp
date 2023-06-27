package lcp

import "fmt"

// MsgCode is the LCP message Code
type MsgCode uint8

// LCP message codes
const (
	CodeConfigureRequest MsgCode = 1
	CodeConfigureAck     MsgCode = 2
	CodeConfigureNak     MsgCode = 3
	CodeConfigureReject  MsgCode = 4
	CodeTerminateRequest MsgCode = 5
	CodeTerminateAck     MsgCode = 6
	CodeCodeReject       MsgCode = 7
	CodeProtocolReject   MsgCode = 8
	CodeEchoRequest      MsgCode = 9
	CodeEchoReply        MsgCode = 10
	CodeDiscardRequest   MsgCode = 11
)

func (code MsgCode) String() string {
	switch code {
	case CodeConfigureRequest:
		return "ConfReq"
	case CodeConfigureAck:
		return "ConfACK"
	case CodeConfigureNak:
		return "ConfNak"
	case CodeConfigureReject:
		return "ConfReject"
	case CodeTerminateRequest:
		return "TermReq"
	case CodeTerminateAck:
		return "TermACK"
	case CodeCodeReject:
		return "CodeReject"
	case CodeProtocolReject:
		return "ProtoReject"
	case CodeEchoRequest:
		return "EchoReq"
	case CodeEchoReply:
		return "EchoReply"
	case CodeDiscardRequest:
		return "DiscardReq"

	}
	return "unknown"
}

// LCPOptionType is the LCP option type
type LCPOptionType uint8

// LCP option types
const (
	OpTypeMaximumReceiveUnit                LCPOptionType = 1
	OpTypeAuthenticationProtocol            LCPOptionType = 3
	OpTypeQualityProtocol                   LCPOptionType = 4
	OpTypeMagicNumber                       LCPOptionType = 5
	OpTypeProtocolFieldCompression          LCPOptionType = 7
	OpTypeAddressandControlFieldCompression LCPOptionType = 8
)

func (op LCPOptionType) String() string {
	switch op {
	case OpTypeMaximumReceiveUnit:
		return "MRU"
	case OpTypeAuthenticationProtocol:
		return "AuthProto"
	case OpTypeQualityProtocol:
		return "QualityProto"
	case OpTypeMagicNumber:
		return "MagicNum"
	case OpTypeProtocolFieldCompression:
		return "ProtoFieldComp"
	case OpTypeAddressandControlFieldCompression:
		return "AddContrlFieldComp"
	}
	return fmt.Sprintf("unknown (%d)", uint8(op))
}

// State is the the LCP protocl state
type State uint32

// LCP protocol state as defined in RFC1661
const (
	StateInitial State = iota
	StateStarting
	StateClosed
	StateStopped
	StateClosing
	StateStopping
	StateReqSent
	StateAckRcvd
	StateAckSent
	StateOpened
	StateEchoReqSent
)

func (s State) String() string {
	switch s {
	case StateInitial:
		return "Initial"
	case StateStarting:
		return "Starting"
	case StateClosed:
		return "Closed"
	case StateStopped:
		return "Stopped"
	case StateClosing:
		return "Closing"
	case StateStopping:
		return "Stopping"
	case StateReqSent:
		return "ReqSent"
	case StateAckRcvd:
		return "AckRcvd"
	case StateAckSent:
		return "AckSent"
	case StateOpened:
		return "Opened"
	case StateEchoReqSent:
		return "EchoReqSent"
	}
	return fmt.Sprintf("unknow (%d)", s)
}

// CHAPAuthAlg is the auth alg of CHAP
type CHAPAuthAlg uint8

// list of CHAP alg
const (
	AlgNone            CHAPAuthAlg = 0
	AlgCHAPwithMD5     CHAPAuthAlg = 5
	AlgSHA1            CHAPAuthAlg = 6
	AlgCHAPwithSHA256  CHAPAuthAlg = 7
	AlgCHAPwithSHA3256 CHAPAuthAlg = 8
	AlgMSCHAP          CHAPAuthAlg = 128
	AlgMSCHAP2         CHAPAuthAlg = 129
)

func (alg CHAPAuthAlg) String() string {
	switch alg {
	case AlgNone:
		return ""
	case AlgCHAPwithMD5:
		return "AlgCHAPwithMD5"
	case AlgSHA1:
		return "AlgSHA1"
	case AlgCHAPwithSHA256:
		return "AlgCHAPwithSHA256"
	case AlgCHAPwithSHA3256:
		return "AlgCHAPwithSHA3256"
	case AlgMSCHAP:
		return "AlgMSCHAP"
	case AlgMSCHAP2:
		return "AlgMSCHAP2"
	}
	return fmt.Sprintf("unknown (%x)", uint8(alg))
}

// LayerNotifyEvent is the tlu/tld/tls/tlf event defined in RFC1661
type LayerNotifyEvent uint8

// list of LayerNotifyEvent
const (
	LCPLayerNotifyUp LayerNotifyEvent = iota
	LCPLayerNotifyDown
	LCPLayerNotifyStarted
	LCPLayerNotifyFinished
)

func (n LayerNotifyEvent) String() string {
	switch n {
	case LCPLayerNotifyUp:
		return "up"
	case LCPLayerNotifyDown:
		return "down"
	case LCPLayerNotifyStarted:
		return "started"
	case LCPLayerNotifyFinished:
		return "finished"
	}
	return fmt.Sprintf("unknown (%d)", n)
}

// IPCPOptionType is the option type for IPCP
type IPCPOptionType uint8

// list of IPCP option type
const (
	OpIPAddresses                IPCPOptionType = 1
	OpIPCompressionProtocol      IPCPOptionType = 2
	OpIPAddress                  IPCPOptionType = 3
	OpMobileIPv4                 IPCPOptionType = 4
	OpPrimaryDNSServerAddress    IPCPOptionType = 129
	OpPrimaryNBNSServerAddress   IPCPOptionType = 130
	OpSecondaryDNSServerAddress  IPCPOptionType = 131
	OpSecondaryNBNSServerAddress IPCPOptionType = 132
)

func (o IPCPOptionType) String() string {
	switch o {
	case OpIPAddresses:
		return "IPAddresses"
	case OpIPCompressionProtocol:
		return "IPCompressionProtocol"
	case OpIPAddress:
		return "IPAddress"
	case OpMobileIPv4:
		return "MobileIPv4"
	case OpPrimaryDNSServerAddress:
		return "PrimaryDNSServerAddress"
	case OpPrimaryNBNSServerAddress:
		return "PrimaryNBNSServerAddress"
	case OpSecondaryDNSServerAddress:
		return "SecondaryDNSServerAddress"
	case OpSecondaryNBNSServerAddress:
		return "SecondaryNBNSServerAddress"
	}
	return fmt.Sprintf("unknown (%d)", o)
}

// PPPProtocolNumber is the PPP protocol number
type PPPProtocolNumber uint16

// list of PPP protocol number
const (
	ProtoNone                                        PPPProtocolNumber = 0
	ProtoPAD                                         PPPProtocolNumber = 0x1
	ProtoIPv4                                        PPPProtocolNumber = 0x21
	ProtoIPv6                                        PPPProtocolNumber = 0x57
	ProtoLCP                                         PPPProtocolNumber = 0xc021
	ProtoPAP                                         PPPProtocolNumber = 0xc023
	ProtoCHAP                                        PPPProtocolNumber = 0xc223
	ProtoEAP                                         PPPProtocolNumber = 0xc227
	ProtoIPCP                                        PPPProtocolNumber = 0x8021
	ProtoIPv6CP                                      PPPProtocolNumber = 0x8057
	ProtoROHCsmallCID                                PPPProtocolNumber = 0x3
	ProtoROHClargeCID                                PPPProtocolNumber = 0x5
	ProtoOSINetworkLayer                             PPPProtocolNumber = 0x23
	ProtoXeroxNSIDP                                  PPPProtocolNumber = 0x25
	ProtoDECnetPhaseIV                               PPPProtocolNumber = 0x27
	ProtoAppletalk                                   PPPProtocolNumber = 0x29
	ProtoNovellIPX                                   PPPProtocolNumber = 0x002b
	ProtoVanJacobsonCompressedTCPIP                  PPPProtocolNumber = 0x002d
	ProtoVanJacobsonUncompressedTCPIP                PPPProtocolNumber = 0x002f
	ProtoBridgingPDU                                 PPPProtocolNumber = 0x31
	ProtoStreamProtocol                              PPPProtocolNumber = 0x33
	ProtoBanyanVines                                 PPPProtocolNumber = 0x35
	ProtoUnassigned                                  PPPProtocolNumber = 0x37
	ProtoAppleTalkEDDP                               PPPProtocolNumber = 0x39
	ProtoAppleTalkSmartBuffered                      PPPProtocolNumber = 0x003b
	ProtoMultiLink                                   PPPProtocolNumber = 0x003d
	ProtoNETBIOSFraming                              PPPProtocolNumber = 0x003f
	ProtoCiscoSystems                                PPPProtocolNumber = 0x41
	ProtoAscomTimeplex                               PPPProtocolNumber = 0x43
	ProtoFujitsuLinkBackupandLoadBalancing           PPPProtocolNumber = 0x45
	ProtoDCARemoteLan                                PPPProtocolNumber = 0x47
	ProtoSerialDataTransportProtocol                 PPPProtocolNumber = 0x49
	ProtoSNAover802                                  PPPProtocolNumber = 0x004b
	ProtoSNA                                         PPPProtocolNumber = 0x004d
	ProtoIPv6HeaderCompression                       PPPProtocolNumber = 0x004f
	ProtoKNXBridgingData                             PPPProtocolNumber = 0x51
	ProtoEncryption                                  PPPProtocolNumber = 0x53
	ProtoIndividualLinkEncryption                    PPPProtocolNumber = 0x55
	ProtoPPPMuxing                                   PPPProtocolNumber = 0x59
	ProtoVendorSpecificNetworkProtocol               PPPProtocolNumber = 0x005b
	ProtoTRILLNetworkProtocol                        PPPProtocolNumber = 0x005d
	ProtoRTPIPHCFullHeader                           PPPProtocolNumber = 0x61
	ProtoRTPIPHCCompressedTCP                        PPPProtocolNumber = 0x63
	ProtoRTPIPHCCompressedNonTCP                     PPPProtocolNumber = 0x65
	ProtoRTPIPHCCompressedUDP8                       PPPProtocolNumber = 0x67
	ProtoRTPIPHCCompressedRTP8                       PPPProtocolNumber = 0x69
	ProtoStampedeBridging                            PPPProtocolNumber = 0x006f
	ProtoMPPlus                                      PPPProtocolNumber = 0x73
	ProtoNTCITSIPI                                   PPPProtocolNumber = 0x00c1
	ProtoSinglelinkcompressioninmultilink            PPPProtocolNumber = 0x00fb
	ProtoCompresseddatagram                          PPPProtocolNumber = 0x00fd
	ProtoHelloPackets8021d                           PPPProtocolNumber = 0x201
	ProtoIBMSourceRoutingBPDU                        PPPProtocolNumber = 0x203
	ProtoDECLANBridge100SpanningTree                 PPPProtocolNumber = 0x205
	ProtoCiscoDiscoveryProtocol                      PPPProtocolNumber = 0x207
	ProtoNetcsTwinRouting                            PPPProtocolNumber = 0x209
	ProtoSTPScheduledTransferProtocol                PPPProtocolNumber = 0x020b
	ProtoEDPExtremeDiscoveryProtocol                 PPPProtocolNumber = 0x020d
	ProtoOpticalSupervisoryChannelProtocol           PPPProtocolNumber = 0x211
	ProtoOpticalSupervisoryChannelProtocolAlias      PPPProtocolNumber = 0x213
	ProtoLuxcom                                      PPPProtocolNumber = 0x231
	ProtoSigmaNetworkSystems                         PPPProtocolNumber = 0x233
	ProtoAppleClientServerProtocol                   PPPProtocolNumber = 0x235
	ProtoMPLSUnicast                                 PPPProtocolNumber = 0x281
	ProtoMPLSMulticast                               PPPProtocolNumber = 0x283
	ProtoIEEEp12844standarddatapackets               PPPProtocolNumber = 0x285
	ProtoETSITETRANetworkProtocolType1               PPPProtocolNumber = 0x287
	ProtoMultichannelFlowTreatmentProtocol           PPPProtocolNumber = 0x289
	ProtoRTPIPHCCompressedTCPNoDelta                 PPPProtocolNumber = 0x2063
	ProtoRTPIPHCContextState                         PPPProtocolNumber = 0x2065
	ProtoRTPIPHCCompressedUDP16                      PPPProtocolNumber = 0x2067
	ProtoRTPIPHCCompressedRTP16                      PPPProtocolNumber = 0x2069
	ProtoCrayCommunicationsControlProtocol           PPPProtocolNumber = 0x4001
	ProtoCDPDMobileNetworkRegistrationProtocol       PPPProtocolNumber = 0x4003
	ProtoExpandacceleratorprotocol                   PPPProtocolNumber = 0x4005
	ProtoODSICPNCP                                   PPPProtocolNumber = 0x4007
	ProtoDOCSISDLL                                   PPPProtocolNumber = 0x4009
	ProtoCetaceanNetworkDetectionProtocol            PPPProtocolNumber = 0x400B
	ProtoStackerLZS                                  PPPProtocolNumber = 0x4021
	ProtoRefTekProtocol                              PPPProtocolNumber = 0x4023
	ProtoFibreChannel                                PPPProtocolNumber = 0x4025
	ProtoOpenDOF                                     PPPProtocolNumber = 0x4027
	ProtoVendorSpecificProtocol                      PPPProtocolNumber = 0x405b
	ProtoTRILLLinkStateProtocol                      PPPProtocolNumber = 0x405d
	ProtoOSINetworkLayerControlProtocol              PPPProtocolNumber = 0x8023
	ProtoXeroxNSIDPControlProtocol                   PPPProtocolNumber = 0x8025
	ProtoDECnetPhaseIVControlProtocol                PPPProtocolNumber = 0x8027
	ProtoAppletalkControlProtocol                    PPPProtocolNumber = 0x8029
	ProtoNovellIPXControlProtocol                    PPPProtocolNumber = 0x802b
	ProtoBridgingNCP                                 PPPProtocolNumber = 0x8031
	ProtoStreamProtocolControlProtocol               PPPProtocolNumber = 0x8033
	ProtoBanyanVinesControlProtocol                  PPPProtocolNumber = 0x8035
	ProtoMultiLinkControlProtocol                    PPPProtocolNumber = 0x803d
	ProtoNETBIOSFramingControlProtocol               PPPProtocolNumber = 0x803f
	ProtoCiscoSystemsControlProtocol                 PPPProtocolNumber = 0x8041
	ProtoAscomTimeplexAlias                          PPPProtocolNumber = 0x8043
	ProtoFujitsuLBLBControlProtocol                  PPPProtocolNumber = 0x8045
	ProtoDCARemoteLanNetworkControlProtocol          PPPProtocolNumber = 0x8047
	ProtoSerialDataControlProtocol                   PPPProtocolNumber = 0x8049
	ProtoSNAover802Control                           PPPProtocolNumber = 0x804b
	ProtoSNAControlProtocol                          PPPProtocolNumber = 0x804d
	ProtoIP6HeaderCompressionControlProtocol         PPPProtocolNumber = 0x804f
	ProtoKNXBridgingControlProtocol                  PPPProtocolNumber = 0x8051
	ProtoEncryptionControlProtocol                   PPPProtocolNumber = 0x8053
	ProtoIndividualLinkEncryptionControlProtocol     PPPProtocolNumber = 0x8055
	ProtoPPPMuxingControlProtocol                    PPPProtocolNumber = 0x8059
	ProtoVendorSpecificNetworkControlProtocol        PPPProtocolNumber = 0x805b
	ProtoTRILLNetworkControlProtocol                 PPPProtocolNumber = 0x805d
	ProtoStampedeBridgingControlProtocol             PPPProtocolNumber = 0x806f
	ProtoMPPlusControlProtocol                       PPPProtocolNumber = 0x8073
	ProtoNTCITSIPIControlProtocol                    PPPProtocolNumber = 0x80c1
	Protosinglelinkcompressioninmultilinkcontrol     PPPProtocolNumber = 0x80fb
	ProtoCompressionControlProtocol                  PPPProtocolNumber = 0x80fd
	ProtoCiscoDiscoveryProtocolControl               PPPProtocolNumber = 0x8207
	ProtoNetcsTwinRoutingAlias                       PPPProtocolNumber = 0x8209
	ProtoSTPControlProtocol                          PPPProtocolNumber = 0x820b
	ProtoEDPCPExtremeDiscoveryProtocolCtrlPrtcl      PPPProtocolNumber = 0x820d
	ProtoAppleClientServerProtocolControl            PPPProtocolNumber = 0x8235
	ProtoMPLSCP                                      PPPProtocolNumber = 0x8281
	ProtoIEEEp12844standardProtocolControl           PPPProtocolNumber = 0x8285
	ProtoETSITETRATNP1ControlProtocol                PPPProtocolNumber = 0x8287
	ProtoMultichannelFlowTreatmentProtocolAlias      PPPProtocolNumber = 0x8289
	ProtoLinkQualityReport                           PPPProtocolNumber = 0xc025
	ProtoShivaPasswordAuthenticationProtocol         PPPProtocolNumber = 0xc027
	ProtoCallBackControlProtocol                     PPPProtocolNumber = 0xc029
	ProtoBACPBandwidthAllocationControlProtocolAlias PPPProtocolNumber = 0xc02b
	ProtoBAP                                         PPPProtocolNumber = 0xc02d
	ProtoVendorSpecificAuthenticationProtocol        PPPProtocolNumber = 0xc05b
	ProtoContainerControlProtocol                    PPPProtocolNumber = 0xc081
	ProtoRSAAuthenticationProtocol                   PPPProtocolNumber = 0xc225
	ProtoMitsubishiSecurityInfoExchPtcl              PPPProtocolNumber = 0xc229
	ProtoStampedeBridgingAuthorizationProtocol       PPPProtocolNumber = 0xc26f
	ProtoProprietaryAuthenticationProtocol           PPPProtocolNumber = 0xc281
	ProtoProprietaryAuthenticationProtocolAlias      PPPProtocolNumber = 0xc283
	ProtoProprietaryNodeIDAuthenticationProtocol     PPPProtocolNumber = 0xc481
)

// func (proto PPPProtocolNumber) String() string {
// 	switch proto {
// 	case ProtoPAD:
// 		return "PADDING"
// 	case ProtoIPv4:
// 		return "IPv4"
// 	case ProtoIPv6:
// 		return "IPv6"
// 	case ProtoLCP:
// 		return "LCP"
// 	case ProtoPAP:
// 		return "PAP"
// 	case ProtoCHAP:
// 		return "CHAP"
// 	case ProtoEAP:
// 		return "EAP"
// 	case ProtoIPCP:
// 		return "IPCP"
// 	case ProtoIPv6CP:
// 		return "IPv6CP"
// 	case ProtoROHCsmallCID:
// 		return "ROHCsmallCID"
// 	case ProtoROHClargeCID:
// 		return "ROHClargeCID"
// 	case ProtoOSINetworkLayer:
// 		return "OSINetworkLayer"
// 	case ProtoXeroxNSIDP:
// 		return "XeroxNSIDP"
// 	case ProtoDECnetPhaseIV:
// 		return "DECnetPhaseIV"
// 	case ProtoAppletalk:
// 		return "Appletalk"
// 	case ProtoNovellIPX:
// 		return "NovellIPX"
// 	case ProtoVanJacobsonCompressedTCPIP:
// 		return "VanJacobsonCompressedTCPIP"
// 	case ProtoVanJacobsonUncompressedTCPIP:
// 		return "VanJacobsonUncompressedTCPIP"
// 	case ProtoBridgingPDU:
// 		return "BridgingPDU"
// 	case ProtoStreamProtocol:
// 		return "StreamProtocol"
// 	case ProtoBanyanVines:
// 		return "BanyanVines"
// 	case ProtoUnassigned:
// 		return "Unassigned"
// 	case ProtoAppleTalkEDDP:
// 		return "AppleTalkEDDP"
// 	case ProtoAppleTalkSmartBuffered:
// 		return "AppleTalkSmartBuffered"
// 	case ProtoMultiLink:
// 		return "MultiLink"
// 	case ProtoNETBIOSFraming:
// 		return "NETBIOSFraming"
// 	case ProtoCiscoSystems:
// 		return "CiscoSystems"
// 	case ProtoAscomTimeplex:
// 		return "AscomTimeplex"
// 	case ProtoFujitsuLinkBackupandLoadBalancing:
// 		return "FujitsuLinkBackupandLoadBalancing"
// 	case ProtoDCARemoteLan:
// 		return "DCARemoteLan"
// 	case ProtoSerialDataTransportProtocol:
// 		return "SerialDataTransportProtocol"
// 	case ProtoSNAover802:
// 		return "SNAover802"
// 	case ProtoSNA:
// 		return "SNA"
// 	case ProtoIPv6HeaderCompression:
// 		return "IPv6HeaderCompression"
// 	case ProtoKNXBridgingData:
// 		return "KNXBridgingData"
// 	case ProtoEncryption:
// 		return "Encryption"
// 	case ProtoIndividualLinkEncryption:
// 		return "IndividualLinkEncryption"
// 	case ProtoPPPMuxing:
// 		return "PPPMuxing"
// 	case ProtoVendorSpecificNetworkProtocol:
// 		return "VendorSpecificNetworkProtocol"
// 	case ProtoTRILLNetworkProtocol:
// 		return "TRILLNetworkProtocol"
// 	case ProtoRTPIPHCFullHeader:
// 		return "RTPIPHCFullHeader"
// 	case ProtoRTPIPHCCompressedTCP:
// 		return "RTPIPHCCompressedTCP"
// 	case ProtoRTPIPHCCompressedNonTCP:
// 		return "RTPIPHCCompressedNonTCP"
// 	case ProtoRTPIPHCCompressedUDP8:
// 		return "RTPIPHCCompressedUDP8"
// 	case ProtoRTPIPHCCompressedRTP8:
// 		return "RTPIPHCCompressedRTP8"
// 	case ProtoStampedeBridging:
// 		return "StampedeBridging"
// 	case ProtoMPPlus:
// 		return "MPPlus"
// 	case ProtoNTCITSIPI:
// 		return "NTCITSIPI"
// 	case ProtoSinglelinkcompressioninmultilink:
// 		return "Singlelinkcompressioninmultilink"
// 	case ProtoCompresseddatagram:
// 		return "Compresseddatagram"
// 	case ProtoHelloPackets8021d:
// 		return "HelloPackets8021d"
// 	case ProtoIBMSourceRoutingBPDU:
// 		return "IBMSourceRoutingBPDU"
// 	case ProtoDECLANBridge100SpanningTree:
// 		return "DECLANBridge100SpanningTree"
// 	case ProtoCiscoDiscoveryProtocol:
// 		return "CiscoDiscoveryProtocol"
// 	case ProtoNetcsTwinRouting:
// 		return "NetcsTwinRouting"
// 	case ProtoSTPScheduledTransferProtocol:
// 		return "STPScheduledTransferProtocol"
// 	case ProtoEDPExtremeDiscoveryProtocol:
// 		return "EDPExtremeDiscoveryProtocol"
// 	case ProtoOpticalSupervisoryChannelProtocol:
// 		return "OpticalSupervisoryChannelProtocol"
// 	case ProtoOpticalSupervisoryChannelProtocolAlias:
// 		return "OpticalSupervisoryChannelProtocolAlias"
// 	case ProtoLuxcom:
// 		return "Luxcom"
// 	case ProtoSigmaNetworkSystems:
// 		return "SigmaNetworkSystems"
// 	case ProtoAppleClientServerProtocol:
// 		return "AppleClientServerProtocol"
// 	case ProtoMPLSUnicast:
// 		return "MPLSUnicast"
// 	case ProtoMPLSMulticast:
// 		return "MPLSMulticast"
// 	case ProtoIEEEp12844standarddatapackets:
// 		return "IEEEp12844standarddatapackets"
// 	case ProtoETSITETRANetworkProtocolType1:
// 		return "ETSITETRANetworkProtocolType1"
// 	case ProtoMultichannelFlowTreatmentProtocol:
// 		return "MultichannelFlowTreatmentProtocol"
// 	case ProtoRTPIPHCCompressedTCPNoDelta:
// 		return "RTPIPHCCompressedTCPNoDelta"
// 	case ProtoRTPIPHCContextState:
// 		return "RTPIPHCContextState"
// 	case ProtoRTPIPHCCompressedUDP16:
// 		return "RTPIPHCCompressedUDP16"
// 	case ProtoRTPIPHCCompressedRTP16:
// 		return "RTPIPHCCompressedRTP16"
// 	case ProtoCrayCommunicationsControlProtocol:
// 		return "CrayCommunicationsControlProtocol"
// 	case ProtoCDPDMobileNetworkRegistrationProtocol:
// 		return "CDPDMobileNetworkRegistrationProtocol"
// 	case ProtoExpandacceleratorprotocol:
// 		return "Expandacceleratorprotocol"
// 	case ProtoODSICPNCP:
// 		return "ODSICPNCP"
// 	case ProtoDOCSISDLL:
// 		return "DOCSISDLL"
// 	case ProtoCetaceanNetworkDetectionProtocol:
// 		return "CetaceanNetworkDetectionProtocol"
// 	case ProtoStackerLZS:
// 		return "StackerLZS"
// 	case ProtoRefTekProtocol:
// 		return "RefTekProtocol"
// 	case ProtoFibreChannel:
// 		return "FibreChannel"
// 	case ProtoOpenDOF:
// 		return "OpenDOF"
// 	case ProtoVendorSpecificProtocol:
// 		return "VendorSpecificProtocol"
// 	case ProtoTRILLLinkStateProtocol:
// 		return "TRILLLinkStateProtocol"
// 	case ProtoOSINetworkLayerControlProtocol:
// 		return "OSINetworkLayerControlProtocol"
// 	case ProtoXeroxNSIDPControlProtocol:
// 		return "XeroxNSIDPControlProtocol"
// 	case ProtoDECnetPhaseIVControlProtocol:
// 		return "DECnetPhaseIVControlProtocol"
// 	case ProtoAppletalkControlProtocol:
// 		return "AppletalkControlProtocol"
// 	case ProtoNovellIPXControlProtocol:
// 		return "NovellIPXControlProtocol"
// 	case ProtoBridgingNCP:
// 		return "BridgingNCP"
// 	case ProtoStreamProtocolControlProtocol:
// 		return "StreamProtocolControlProtocol"
// 	case ProtoBanyanVinesControlProtocol:
// 		return "BanyanVinesControlProtocol"
// 	case ProtoMultiLinkControlProtocol:
// 		return "MultiLinkControlProtocol"
// 	case ProtoNETBIOSFramingControlProtocol:
// 		return "NETBIOSFramingControlProtocol"
// 	case ProtoCiscoSystemsControlProtocol:
// 		return "CiscoSystemsControlProtocol"
// 	case ProtoAscomTimeplexAlias:
// 		return "AscomTimeplexAlias"
// 	case ProtoFujitsuLBLBControlProtocol:
// 		return "FujitsuLBLBControlProtocol"
// 	case ProtoDCARemoteLanNetworkControlProtocol:
// 		return "DCARemoteLanNetworkControlProtocol"
// 	case ProtoSerialDataControlProtocol:
// 		return "SerialDataControlProtocol"
// 	case ProtoSNAover802Control:
// 		return "SNAover802Control"
// 	case ProtoSNAControlProtocol:
// 		return "SNAControlProtocol"
// 	case ProtoIP6HeaderCompressionControlProtocol:
// 		return "IP6HeaderCompressionControlProtocol"
// 	case ProtoKNXBridgingControlProtocol:
// 		return "KNXBridgingControlProtocol"
// 	case ProtoEncryptionControlProtocol:
// 		return "EncryptionControlProtocol"
// 	case ProtoIndividualLinkEncryptionControlProtocol:
// 		return "IndividualLinkEncryptionControlProtocol"
// 	case ProtoPPPMuxingControlProtocol:
// 		return "PPPMuxingControlProtocol"
// 	case ProtoVendorSpecificNetworkControlProtocol:
// 		return "VendorSpecificNetworkControlProtocol"
// 	case ProtoTRILLNetworkControlProtocol:
// 		return "TRILLNetworkControlProtocol"
// 	case ProtoStampedeBridgingControlProtocol:
// 		return "StampedeBridgingControlProtocol"
// 	case ProtoMPPlusControlProtocol:
// 		return "MPPlusControlProtocol"
// 	case ProtoNTCITSIPIControlProtocol:
// 		return "NTCITSIPIControlProtocol"
// 	case Protosinglelinkcompressioninmultilinkcontrol:
// 		return "singlelinkcompressioninmultilinkcontrol"
// 	case ProtoCompressionControlProtocol:
// 		return "CompressionControlProtocol"
// 	case ProtoCiscoDiscoveryProtocolControl:
// 		return "CiscoDiscoveryProtocolControl"
// 	case ProtoNetcsTwinRoutingAlias:
// 		return "NetcsTwinRoutingAlias"
// 	case ProtoSTPControlProtocol:
// 		return "STPControlProtocol"
// 	case ProtoEDPCPExtremeDiscoveryProtocolCtrlPrtcl:
// 		return "EDPCPExtremeDiscoveryProtocolCtrlPrtcl"
// 	case ProtoAppleClientServerProtocolControl:
// 		return "AppleClientServerProtocolControl"
// 	case ProtoMPLSCP:
// 		return "MPLSCP"
// 	case ProtoIEEEp12844standardProtocolControl:
// 		return "IEEEp12844standardProtocolControl"
// 	case ProtoETSITETRATNP1ControlProtocol:
// 		return "ETSITETRATNP1ControlProtocol"
// 	case ProtoMultichannelFlowTreatmentProtocolAlias:
// 		return "MultichannelFlowTreatmentProtocolAlias"
// 	case ProtoLinkQualityReport:
// 		return "LinkQualityReport"
// 	case ProtoShivaPasswordAuthenticationProtocol:
// 		return "ShivaPasswordAuthenticationProtocol"
// 	case ProtoCallBackControlProtocol:
// 		return "CallBackControlProtocol"
// 	case ProtoBACPBandwidthAllocationControlProtocolAlias:
// 		return "BACPBandwidthAllocationControlProtocolAlias"
// 	case ProtoBAP:
// 		return "BAP"
// 	case ProtoVendorSpecificAuthenticationProtocol:
// 		return "VendorSpecificAuthenticationProtocol"
// 	case ProtoContainerControlProtocol:
// 		return "ContainerControlProtocol"
// 	case ProtoRSAAuthenticationProtocol:
// 		return "RSAAuthenticationProtocol"
// 	case ProtoMitsubishiSecurityInfoExchPtcl:
// 		return "MitsubishiSecurityInfoExchPtcl"
// 	case ProtoStampedeBridgingAuthorizationProtocol:
// 		return "StampedeBridgingAuthorizationProtocol"
// 	case ProtoProprietaryAuthenticationProtocol:
// 		return "ProprietaryAuthenticationProtocol"
// 	case ProtoProprietaryAuthenticationProtocolAlias:
// 		return "ProprietaryAuthenticationProtocolAlias"
// 	case ProtoProprietaryNodeIDAuthenticationProtocol:
// 		return "ProprietaryNodeIDAuthenticationProtocol"
// 	}
// 	return fmt.Sprintf("unknown (%x)", uint16(proto))
// }

// IPCP6OptionType is the option type for IPv6CP
type IPCP6OptionType uint8

// list of IPv6CP option type
const (
	IP6CPOpIPv6CompressionProtocol IPCP6OptionType = 0x2
	IP6CPOpInterfaceIdentifier     IPCP6OptionType = 0x1
)

func (code IPCP6OptionType) String() string {
	switch code {

	case IP6CPOpIPv6CompressionProtocol:
		return "IPv6CompressionProtocol"

	case IP6CPOpInterfaceIdentifier:
		return "InterfaceIdentifier"

	}
	return fmt.Sprintf("unknown (%d)", code)
}
