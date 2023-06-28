package lcp

import "fmt"

func (val PPPProtocolNumber) String() string {
	r,err:=val.MarshalText()
	if err!=nil {
		return fmt.Sprint(err)
	}
	return string(r)
}

func (val PPPProtocolNumber) MarshalText() (text []byte, err error) {
	switch val {
	 
	case ProtoAppleClientServerProtocol:
		return []byte("AppleClientServerProtocol"),nil
	 
	case ProtoAppleClientServerProtocolControl:
		return []byte("AppleClientServerProtocolControl"),nil
	 
	case ProtoAppleTalkEDDP:
		return []byte("AppleTalkEDDP"),nil
	 
	case ProtoAppleTalkSmartBuffered:
		return []byte("AppleTalkSmartBuffered"),nil
	 
	case ProtoAppletalk:
		return []byte("Appletalk"),nil
	 
	case ProtoAppletalkControlProtocol:
		return []byte("AppletalkControlProtocol"),nil
	 
	case ProtoAscomTimeplex:
		return []byte("AscomTimeplex"),nil
	 
	case ProtoAscomTimeplexAlias:
		return []byte("AscomTimeplexAlias"),nil
	 
	case ProtoBACPBandwidthAllocationControlProtocolAlias:
		return []byte("BACPBandwidthAllocationControlProtocolAlias"),nil
	 
	case ProtoBAP:
		return []byte("BAP"),nil
	 
	case ProtoBanyanVines:
		return []byte("BanyanVines"),nil
	 
	case ProtoBanyanVinesControlProtocol:
		return []byte("BanyanVinesControlProtocol"),nil
	 
	case ProtoBridgingNCP:
		return []byte("BridgingNCP"),nil
	 
	case ProtoBridgingPDU:
		return []byte("BridgingPDU"),nil
	 
	case ProtoCDPDMobileNetworkRegistrationProtocol:
		return []byte("CDPDMobileNetworkRegistrationProtocol"),nil
	 
	case ProtoCHAP:
		return []byte("CHAP"),nil
	 
	case ProtoCallBackControlProtocol:
		return []byte("CallBackControlProtocol"),nil
	 
	case ProtoCetaceanNetworkDetectionProtocol:
		return []byte("CetaceanNetworkDetectionProtocol"),nil
	 
	case ProtoCiscoDiscoveryProtocol:
		return []byte("CiscoDiscoveryProtocol"),nil
	 
	case ProtoCiscoDiscoveryProtocolControl:
		return []byte("CiscoDiscoveryProtocolControl"),nil
	 
	case ProtoCiscoSystems:
		return []byte("CiscoSystems"),nil
	 
	case ProtoCiscoSystemsControlProtocol:
		return []byte("CiscoSystemsControlProtocol"),nil
	 
	case ProtoCompresseddatagram:
		return []byte("Compresseddatagram"),nil
	 
	case ProtoCompressionControlProtocol:
		return []byte("CompressionControlProtocol"),nil
	 
	case ProtoContainerControlProtocol:
		return []byte("ContainerControlProtocol"),nil
	 
	case ProtoCrayCommunicationsControlProtocol:
		return []byte("CrayCommunicationsControlProtocol"),nil
	 
	case ProtoDCARemoteLan:
		return []byte("DCARemoteLan"),nil
	 
	case ProtoDCARemoteLanNetworkControlProtocol:
		return []byte("DCARemoteLanNetworkControlProtocol"),nil
	 
	case ProtoDECLANBridge100SpanningTree:
		return []byte("DECLANBridge100SpanningTree"),nil
	 
	case ProtoDECnetPhaseIV:
		return []byte("DECnetPhaseIV"),nil
	 
	case ProtoDECnetPhaseIVControlProtocol:
		return []byte("DECnetPhaseIVControlProtocol"),nil
	 
	case ProtoDOCSISDLL:
		return []byte("DOCSISDLL"),nil
	 
	case ProtoEAP:
		return []byte("EAP"),nil
	 
	case ProtoEDPCPExtremeDiscoveryProtocolCtrlPrtcl:
		return []byte("EDPCPExtremeDiscoveryProtocolCtrlPrtcl"),nil
	 
	case ProtoEDPExtremeDiscoveryProtocol:
		return []byte("EDPExtremeDiscoveryProtocol"),nil
	 
	case ProtoETSITETRANetworkProtocolType1:
		return []byte("ETSITETRANetworkProtocolType1"),nil
	 
	case ProtoETSITETRATNP1ControlProtocol:
		return []byte("ETSITETRATNP1ControlProtocol"),nil
	 
	case ProtoEncryption:
		return []byte("Encryption"),nil
	 
	case ProtoEncryptionControlProtocol:
		return []byte("EncryptionControlProtocol"),nil
	 
	case ProtoExpandacceleratorprotocol:
		return []byte("Expandacceleratorprotocol"),nil
	 
	case ProtoFibreChannel:
		return []byte("FibreChannel"),nil
	 
	case ProtoFujitsuLBLBControlProtocol:
		return []byte("FujitsuLBLBControlProtocol"),nil
	 
	case ProtoFujitsuLinkBackupandLoadBalancing:
		return []byte("FujitsuLinkBackupandLoadBalancing"),nil
	 
	case ProtoHelloPackets8021d:
		return []byte("HelloPackets8021d"),nil
	 
	case ProtoIBMSourceRoutingBPDU:
		return []byte("IBMSourceRoutingBPDU"),nil
	 
	case ProtoIEEEp12844standardProtocolControl:
		return []byte("IEEEp12844standardProtocolControl"),nil
	 
	case ProtoIEEEp12844standarddatapackets:
		return []byte("IEEEp12844standarddatapackets"),nil
	 
	case ProtoIP6HeaderCompressionControlProtocol:
		return []byte("IP6HeaderCompressionControlProtocol"),nil
	 
	case ProtoIPCP:
		return []byte("IPCP"),nil
	 
	case ProtoIPv4:
		return []byte("IPv4"),nil
	 
	case ProtoIPv6:
		return []byte("IPv6"),nil
	 
	case ProtoIPv6CP:
		return []byte("IPv6CP"),nil
	 
	case ProtoIPv6HeaderCompression:
		return []byte("IPv6HeaderCompression"),nil
	 
	case ProtoIndividualLinkEncryption:
		return []byte("IndividualLinkEncryption"),nil
	 
	case ProtoIndividualLinkEncryptionControlProtocol:
		return []byte("IndividualLinkEncryptionControlProtocol"),nil
	 
	case ProtoKNXBridgingControlProtocol:
		return []byte("KNXBridgingControlProtocol"),nil
	 
	case ProtoKNXBridgingData:
		return []byte("KNXBridgingData"),nil
	 
	case ProtoLCP:
		return []byte("LCP"),nil
	 
	case ProtoLinkQualityReport:
		return []byte("LinkQualityReport"),nil
	 
	case ProtoLuxcom:
		return []byte("Luxcom"),nil
	 
	case ProtoMPLSCP:
		return []byte("MPLSCP"),nil
	 
	case ProtoMPLSMulticast:
		return []byte("MPLSMulticast"),nil
	 
	case ProtoMPLSUnicast:
		return []byte("MPLSUnicast"),nil
	 
	case ProtoMPPlus:
		return []byte("MPPlus"),nil
	 
	case ProtoMPPlusControlProtocol:
		return []byte("MPPlusControlProtocol"),nil
	 
	case ProtoMitsubishiSecurityInfoExchPtcl:
		return []byte("MitsubishiSecurityInfoExchPtcl"),nil
	 
	case ProtoMultiLink:
		return []byte("MultiLink"),nil
	 
	case ProtoMultiLinkControlProtocol:
		return []byte("MultiLinkControlProtocol"),nil
	 
	case ProtoMultichannelFlowTreatmentProtocol:
		return []byte("MultichannelFlowTreatmentProtocol"),nil
	 
	case ProtoMultichannelFlowTreatmentProtocolAlias:
		return []byte("MultichannelFlowTreatmentProtocolAlias"),nil
	 
	case ProtoNETBIOSFraming:
		return []byte("NETBIOSFraming"),nil
	 
	case ProtoNETBIOSFramingControlProtocol:
		return []byte("NETBIOSFramingControlProtocol"),nil
	 
	case ProtoNTCITSIPI:
		return []byte("NTCITSIPI"),nil
	 
	case ProtoNTCITSIPIControlProtocol:
		return []byte("NTCITSIPIControlProtocol"),nil
	 
	case ProtoNetcsTwinRouting:
		return []byte("NetcsTwinRouting"),nil
	 
	case ProtoNetcsTwinRoutingAlias:
		return []byte("NetcsTwinRoutingAlias"),nil
	 
	case ProtoNone:
		return []byte("None"),nil
	 
	case ProtoNovellIPX:
		return []byte("NovellIPX"),nil
	 
	case ProtoNovellIPXControlProtocol:
		return []byte("NovellIPXControlProtocol"),nil
	 
	case ProtoODSICPNCP:
		return []byte("ODSICPNCP"),nil
	 
	case ProtoOSINetworkLayer:
		return []byte("OSINetworkLayer"),nil
	 
	case ProtoOSINetworkLayerControlProtocol:
		return []byte("OSINetworkLayerControlProtocol"),nil
	 
	case ProtoOpenDOF:
		return []byte("OpenDOF"),nil
	 
	case ProtoOpticalSupervisoryChannelProtocol:
		return []byte("OpticalSupervisoryChannelProtocol"),nil
	 
	case ProtoOpticalSupervisoryChannelProtocolAlias:
		return []byte("OpticalSupervisoryChannelProtocolAlias"),nil
	 
	case ProtoPAD:
		return []byte("PAD"),nil
	 
	case ProtoPAP:
		return []byte("PAP"),nil
	 
	case ProtoPPPMuxing:
		return []byte("PPPMuxing"),nil
	 
	case ProtoPPPMuxingControlProtocol:
		return []byte("PPPMuxingControlProtocol"),nil
	 
	case ProtoProprietaryAuthenticationProtocol:
		return []byte("ProprietaryAuthenticationProtocol"),nil
	 
	case ProtoProprietaryAuthenticationProtocolAlias:
		return []byte("ProprietaryAuthenticationProtocolAlias"),nil
	 
	case ProtoProprietaryNodeIDAuthenticationProtocol:
		return []byte("ProprietaryNodeIDAuthenticationProtocol"),nil
	 
	case ProtoROHClargeCID:
		return []byte("ROHClargeCID"),nil
	 
	case ProtoROHCsmallCID:
		return []byte("ROHCsmallCID"),nil
	 
	case ProtoRSAAuthenticationProtocol:
		return []byte("RSAAuthenticationProtocol"),nil
	 
	case ProtoRTPIPHCCompressedNonTCP:
		return []byte("RTPIPHCCompressedNonTCP"),nil
	 
	case ProtoRTPIPHCCompressedRTP16:
		return []byte("RTPIPHCCompressedRTP16"),nil
	 
	case ProtoRTPIPHCCompressedRTP8:
		return []byte("RTPIPHCCompressedRTP8"),nil
	 
	case ProtoRTPIPHCCompressedTCP:
		return []byte("RTPIPHCCompressedTCP"),nil
	 
	case ProtoRTPIPHCCompressedTCPNoDelta:
		return []byte("RTPIPHCCompressedTCPNoDelta"),nil
	 
	case ProtoRTPIPHCCompressedUDP16:
		return []byte("RTPIPHCCompressedUDP16"),nil
	 
	case ProtoRTPIPHCCompressedUDP8:
		return []byte("RTPIPHCCompressedUDP8"),nil
	 
	case ProtoRTPIPHCContextState:
		return []byte("RTPIPHCContextState"),nil
	 
	case ProtoRTPIPHCFullHeader:
		return []byte("RTPIPHCFullHeader"),nil
	 
	case ProtoRefTekProtocol:
		return []byte("RefTekProtocol"),nil
	 
	case ProtoSNA:
		return []byte("SNA"),nil
	 
	case ProtoSNAControlProtocol:
		return []byte("SNAControlProtocol"),nil
	 
	case ProtoSNAover802:
		return []byte("SNAover802"),nil
	 
	case ProtoSNAover802Control:
		return []byte("SNAover802Control"),nil
	 
	case ProtoSTPControlProtocol:
		return []byte("STPControlProtocol"),nil
	 
	case ProtoSTPScheduledTransferProtocol:
		return []byte("STPScheduledTransferProtocol"),nil
	 
	case ProtoSerialDataControlProtocol:
		return []byte("SerialDataControlProtocol"),nil
	 
	case ProtoSerialDataTransportProtocol:
		return []byte("SerialDataTransportProtocol"),nil
	 
	case ProtoShivaPasswordAuthenticationProtocol:
		return []byte("ShivaPasswordAuthenticationProtocol"),nil
	 
	case ProtoSigmaNetworkSystems:
		return []byte("SigmaNetworkSystems"),nil
	 
	case ProtoSinglelinkcompressioninmultilink:
		return []byte("Singlelinkcompressioninmultilink"),nil
	 
	case ProtoStackerLZS:
		return []byte("StackerLZS"),nil
	 
	case ProtoStampedeBridging:
		return []byte("StampedeBridging"),nil
	 
	case ProtoStampedeBridgingAuthorizationProtocol:
		return []byte("StampedeBridgingAuthorizationProtocol"),nil
	 
	case ProtoStampedeBridgingControlProtocol:
		return []byte("StampedeBridgingControlProtocol"),nil
	 
	case ProtoStreamProtocol:
		return []byte("StreamProtocol"),nil
	 
	case ProtoStreamProtocolControlProtocol:
		return []byte("StreamProtocolControlProtocol"),nil
	 
	case ProtoTRILLLinkStateProtocol:
		return []byte("TRILLLinkStateProtocol"),nil
	 
	case ProtoTRILLNetworkControlProtocol:
		return []byte("TRILLNetworkControlProtocol"),nil
	 
	case ProtoTRILLNetworkProtocol:
		return []byte("TRILLNetworkProtocol"),nil
	 
	case ProtoUnassigned:
		return []byte("Unassigned"),nil
	 
	case ProtoVanJacobsonCompressedTCPIP:
		return []byte("VanJacobsonCompressedTCPIP"),nil
	 
	case ProtoVanJacobsonUncompressedTCPIP:
		return []byte("VanJacobsonUncompressedTCPIP"),nil
	 
	case ProtoVendorSpecificAuthenticationProtocol:
		return []byte("VendorSpecificAuthenticationProtocol"),nil
	 
	case ProtoVendorSpecificNetworkControlProtocol:
		return []byte("VendorSpecificNetworkControlProtocol"),nil
	 
	case ProtoVendorSpecificNetworkProtocol:
		return []byte("VendorSpecificNetworkProtocol"),nil
	 
	case ProtoVendorSpecificProtocol:
		return []byte("VendorSpecificProtocol"),nil
	 
	case ProtoXeroxNSIDP:
		return []byte("XeroxNSIDP"),nil
	 
	case ProtoXeroxNSIDPControlProtocol:
		return []byte("XeroxNSIDPControlProtocol"),nil
	 
	case Protosinglelinkcompressioninmultilinkcontrol:
		return []byte("singlelinkcompressioninmultilinkcontrol"),nil
	
	}
	return nil, fmt.Errorf("unknown value %#v", val)
}

func (val *PPPProtocolNumber) UnmarshalText(text []byte) error {
	input := string(text)
	switch input {
	 
	case "AppleClientServerProtocol":
		*val=ProtoAppleClientServerProtocol
	 
	case "AppleClientServerProtocolControl":
		*val=ProtoAppleClientServerProtocolControl
	 
	case "AppleTalkEDDP":
		*val=ProtoAppleTalkEDDP
	 
	case "AppleTalkSmartBuffered":
		*val=ProtoAppleTalkSmartBuffered
	 
	case "Appletalk":
		*val=ProtoAppletalk
	 
	case "AppletalkControlProtocol":
		*val=ProtoAppletalkControlProtocol
	 
	case "AscomTimeplex":
		*val=ProtoAscomTimeplex
	 
	case "AscomTimeplexAlias":
		*val=ProtoAscomTimeplexAlias
	 
	case "BACPBandwidthAllocationControlProtocolAlias":
		*val=ProtoBACPBandwidthAllocationControlProtocolAlias
	 
	case "BAP":
		*val=ProtoBAP
	 
	case "BanyanVines":
		*val=ProtoBanyanVines
	 
	case "BanyanVinesControlProtocol":
		*val=ProtoBanyanVinesControlProtocol
	 
	case "BridgingNCP":
		*val=ProtoBridgingNCP
	 
	case "BridgingPDU":
		*val=ProtoBridgingPDU
	 
	case "CDPDMobileNetworkRegistrationProtocol":
		*val=ProtoCDPDMobileNetworkRegistrationProtocol
	 
	case "CHAP":
		*val=ProtoCHAP
	 
	case "CallBackControlProtocol":
		*val=ProtoCallBackControlProtocol
	 
	case "CetaceanNetworkDetectionProtocol":
		*val=ProtoCetaceanNetworkDetectionProtocol
	 
	case "CiscoDiscoveryProtocol":
		*val=ProtoCiscoDiscoveryProtocol
	 
	case "CiscoDiscoveryProtocolControl":
		*val=ProtoCiscoDiscoveryProtocolControl
	 
	case "CiscoSystems":
		*val=ProtoCiscoSystems
	 
	case "CiscoSystemsControlProtocol":
		*val=ProtoCiscoSystemsControlProtocol
	 
	case "Compresseddatagram":
		*val=ProtoCompresseddatagram
	 
	case "CompressionControlProtocol":
		*val=ProtoCompressionControlProtocol
	 
	case "ContainerControlProtocol":
		*val=ProtoContainerControlProtocol
	 
	case "CrayCommunicationsControlProtocol":
		*val=ProtoCrayCommunicationsControlProtocol
	 
	case "DCARemoteLan":
		*val=ProtoDCARemoteLan
	 
	case "DCARemoteLanNetworkControlProtocol":
		*val=ProtoDCARemoteLanNetworkControlProtocol
	 
	case "DECLANBridge100SpanningTree":
		*val=ProtoDECLANBridge100SpanningTree
	 
	case "DECnetPhaseIV":
		*val=ProtoDECnetPhaseIV
	 
	case "DECnetPhaseIVControlProtocol":
		*val=ProtoDECnetPhaseIVControlProtocol
	 
	case "DOCSISDLL":
		*val=ProtoDOCSISDLL
	 
	case "EAP":
		*val=ProtoEAP
	 
	case "EDPCPExtremeDiscoveryProtocolCtrlPrtcl":
		*val=ProtoEDPCPExtremeDiscoveryProtocolCtrlPrtcl
	 
	case "EDPExtremeDiscoveryProtocol":
		*val=ProtoEDPExtremeDiscoveryProtocol
	 
	case "ETSITETRANetworkProtocolType1":
		*val=ProtoETSITETRANetworkProtocolType1
	 
	case "ETSITETRATNP1ControlProtocol":
		*val=ProtoETSITETRATNP1ControlProtocol
	 
	case "Encryption":
		*val=ProtoEncryption
	 
	case "EncryptionControlProtocol":
		*val=ProtoEncryptionControlProtocol
	 
	case "Expandacceleratorprotocol":
		*val=ProtoExpandacceleratorprotocol
	 
	case "FibreChannel":
		*val=ProtoFibreChannel
	 
	case "FujitsuLBLBControlProtocol":
		*val=ProtoFujitsuLBLBControlProtocol
	 
	case "FujitsuLinkBackupandLoadBalancing":
		*val=ProtoFujitsuLinkBackupandLoadBalancing
	 
	case "HelloPackets8021d":
		*val=ProtoHelloPackets8021d
	 
	case "IBMSourceRoutingBPDU":
		*val=ProtoIBMSourceRoutingBPDU
	 
	case "IEEEp12844standardProtocolControl":
		*val=ProtoIEEEp12844standardProtocolControl
	 
	case "IEEEp12844standarddatapackets":
		*val=ProtoIEEEp12844standarddatapackets
	 
	case "IP6HeaderCompressionControlProtocol":
		*val=ProtoIP6HeaderCompressionControlProtocol
	 
	case "IPCP":
		*val=ProtoIPCP
	 
	case "IPv4":
		*val=ProtoIPv4
	 
	case "IPv6":
		*val=ProtoIPv6
	 
	case "IPv6CP":
		*val=ProtoIPv6CP
	 
	case "IPv6HeaderCompression":
		*val=ProtoIPv6HeaderCompression
	 
	case "IndividualLinkEncryption":
		*val=ProtoIndividualLinkEncryption
	 
	case "IndividualLinkEncryptionControlProtocol":
		*val=ProtoIndividualLinkEncryptionControlProtocol
	 
	case "KNXBridgingControlProtocol":
		*val=ProtoKNXBridgingControlProtocol
	 
	case "KNXBridgingData":
		*val=ProtoKNXBridgingData
	 
	case "LCP":
		*val=ProtoLCP
	 
	case "LinkQualityReport":
		*val=ProtoLinkQualityReport
	 
	case "Luxcom":
		*val=ProtoLuxcom
	 
	case "MPLSCP":
		*val=ProtoMPLSCP
	 
	case "MPLSMulticast":
		*val=ProtoMPLSMulticast
	 
	case "MPLSUnicast":
		*val=ProtoMPLSUnicast
	 
	case "MPPlus":
		*val=ProtoMPPlus
	 
	case "MPPlusControlProtocol":
		*val=ProtoMPPlusControlProtocol
	 
	case "MitsubishiSecurityInfoExchPtcl":
		*val=ProtoMitsubishiSecurityInfoExchPtcl
	 
	case "MultiLink":
		*val=ProtoMultiLink
	 
	case "MultiLinkControlProtocol":
		*val=ProtoMultiLinkControlProtocol
	 
	case "MultichannelFlowTreatmentProtocol":
		*val=ProtoMultichannelFlowTreatmentProtocol
	 
	case "MultichannelFlowTreatmentProtocolAlias":
		*val=ProtoMultichannelFlowTreatmentProtocolAlias
	 
	case "NETBIOSFraming":
		*val=ProtoNETBIOSFraming
	 
	case "NETBIOSFramingControlProtocol":
		*val=ProtoNETBIOSFramingControlProtocol
	 
	case "NTCITSIPI":
		*val=ProtoNTCITSIPI
	 
	case "NTCITSIPIControlProtocol":
		*val=ProtoNTCITSIPIControlProtocol
	 
	case "NetcsTwinRouting":
		*val=ProtoNetcsTwinRouting
	 
	case "NetcsTwinRoutingAlias":
		*val=ProtoNetcsTwinRoutingAlias
	 
	case "None":
		*val=ProtoNone
	 
	case "NovellIPX":
		*val=ProtoNovellIPX
	 
	case "NovellIPXControlProtocol":
		*val=ProtoNovellIPXControlProtocol
	 
	case "ODSICPNCP":
		*val=ProtoODSICPNCP
	 
	case "OSINetworkLayer":
		*val=ProtoOSINetworkLayer
	 
	case "OSINetworkLayerControlProtocol":
		*val=ProtoOSINetworkLayerControlProtocol
	 
	case "OpenDOF":
		*val=ProtoOpenDOF
	 
	case "OpticalSupervisoryChannelProtocol":
		*val=ProtoOpticalSupervisoryChannelProtocol
	 
	case "OpticalSupervisoryChannelProtocolAlias":
		*val=ProtoOpticalSupervisoryChannelProtocolAlias
	 
	case "PAD":
		*val=ProtoPAD
	 
	case "PAP":
		*val=ProtoPAP
	 
	case "PPPMuxing":
		*val=ProtoPPPMuxing
	 
	case "PPPMuxingControlProtocol":
		*val=ProtoPPPMuxingControlProtocol
	 
	case "ProprietaryAuthenticationProtocol":
		*val=ProtoProprietaryAuthenticationProtocol
	 
	case "ProprietaryAuthenticationProtocolAlias":
		*val=ProtoProprietaryAuthenticationProtocolAlias
	 
	case "ProprietaryNodeIDAuthenticationProtocol":
		*val=ProtoProprietaryNodeIDAuthenticationProtocol
	 
	case "ROHClargeCID":
		*val=ProtoROHClargeCID
	 
	case "ROHCsmallCID":
		*val=ProtoROHCsmallCID
	 
	case "RSAAuthenticationProtocol":
		*val=ProtoRSAAuthenticationProtocol
	 
	case "RTPIPHCCompressedNonTCP":
		*val=ProtoRTPIPHCCompressedNonTCP
	 
	case "RTPIPHCCompressedRTP16":
		*val=ProtoRTPIPHCCompressedRTP16
	 
	case "RTPIPHCCompressedRTP8":
		*val=ProtoRTPIPHCCompressedRTP8
	 
	case "RTPIPHCCompressedTCP":
		*val=ProtoRTPIPHCCompressedTCP
	 
	case "RTPIPHCCompressedTCPNoDelta":
		*val=ProtoRTPIPHCCompressedTCPNoDelta
	 
	case "RTPIPHCCompressedUDP16":
		*val=ProtoRTPIPHCCompressedUDP16
	 
	case "RTPIPHCCompressedUDP8":
		*val=ProtoRTPIPHCCompressedUDP8
	 
	case "RTPIPHCContextState":
		*val=ProtoRTPIPHCContextState
	 
	case "RTPIPHCFullHeader":
		*val=ProtoRTPIPHCFullHeader
	 
	case "RefTekProtocol":
		*val=ProtoRefTekProtocol
	 
	case "SNA":
		*val=ProtoSNA
	 
	case "SNAControlProtocol":
		*val=ProtoSNAControlProtocol
	 
	case "SNAover802":
		*val=ProtoSNAover802
	 
	case "SNAover802Control":
		*val=ProtoSNAover802Control
	 
	case "STPControlProtocol":
		*val=ProtoSTPControlProtocol
	 
	case "STPScheduledTransferProtocol":
		*val=ProtoSTPScheduledTransferProtocol
	 
	case "SerialDataControlProtocol":
		*val=ProtoSerialDataControlProtocol
	 
	case "SerialDataTransportProtocol":
		*val=ProtoSerialDataTransportProtocol
	 
	case "ShivaPasswordAuthenticationProtocol":
		*val=ProtoShivaPasswordAuthenticationProtocol
	 
	case "SigmaNetworkSystems":
		*val=ProtoSigmaNetworkSystems
	 
	case "Singlelinkcompressioninmultilink":
		*val=ProtoSinglelinkcompressioninmultilink
	 
	case "StackerLZS":
		*val=ProtoStackerLZS
	 
	case "StampedeBridging":
		*val=ProtoStampedeBridging
	 
	case "StampedeBridgingAuthorizationProtocol":
		*val=ProtoStampedeBridgingAuthorizationProtocol
	 
	case "StampedeBridgingControlProtocol":
		*val=ProtoStampedeBridgingControlProtocol
	 
	case "StreamProtocol":
		*val=ProtoStreamProtocol
	 
	case "StreamProtocolControlProtocol":
		*val=ProtoStreamProtocolControlProtocol
	 
	case "TRILLLinkStateProtocol":
		*val=ProtoTRILLLinkStateProtocol
	 
	case "TRILLNetworkControlProtocol":
		*val=ProtoTRILLNetworkControlProtocol
	 
	case "TRILLNetworkProtocol":
		*val=ProtoTRILLNetworkProtocol
	 
	case "Unassigned":
		*val=ProtoUnassigned
	 
	case "VanJacobsonCompressedTCPIP":
		*val=ProtoVanJacobsonCompressedTCPIP
	 
	case "VanJacobsonUncompressedTCPIP":
		*val=ProtoVanJacobsonUncompressedTCPIP
	 
	case "VendorSpecificAuthenticationProtocol":
		*val=ProtoVendorSpecificAuthenticationProtocol
	 
	case "VendorSpecificNetworkControlProtocol":
		*val=ProtoVendorSpecificNetworkControlProtocol
	 
	case "VendorSpecificNetworkProtocol":
		*val=ProtoVendorSpecificNetworkProtocol
	 
	case "VendorSpecificProtocol":
		*val=ProtoVendorSpecificProtocol
	 
	case "XeroxNSIDP":
		*val=ProtoXeroxNSIDP
	 
	case "XeroxNSIDPControlProtocol":
		*val=ProtoXeroxNSIDPControlProtocol
	 
	case "singlelinkcompressioninmultilinkcontrol":
		*val=Protosinglelinkcompressioninmultilinkcontrol
	
	default:
		return fmt.Errorf("failed to parse %v into PPPProtocolNumber", input)
	}
	return nil
}
	