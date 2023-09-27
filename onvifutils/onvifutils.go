package onvifutils

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/aler9/gortsplib/pkg/url"
	"github.com/sonnt85/gosutils/slogrus"

	"github.com/aler9/gortsplib"
	"github.com/beevik/etree"
	"github.com/lunny/log"
	"github.com/sonnt85/gonvif"
	imaging "github.com/sonnt85/gonvif/Imaging"
	"github.com/sonnt85/gonvif/analytics"
	"github.com/sonnt85/gosutils/gcurl"

	"github.com/sonnt85/gonvif/device"
	"github.com/sonnt85/gonvif/event"
	"github.com/sonnt85/gonvif/gosoap"
	"github.com/sonnt85/gonvif/media"
	"github.com/sonnt85/gonvif/networking"
	"github.com/sonnt85/gonvif/ptz"
	"github.com/sonnt85/gosutils/sexec"
	"github.com/sonnt85/gosutils/sregexp"
	"github.com/sonnt85/gosutils/sutils"
	"github.com/sonnt85/snetutils"
)

var GetOnvifStruct = map[string]map[string]interface{}{
	"imaging":   {"GetServiceCapabilities": &imaging.GetServiceCapabilities{}, "GetImagingSettings": &imaging.GetImagingSettings{}, "SetImagingSettings": &imaging.SetImagingSettings{}, "GetOptions": &imaging.GetOptions{}, "Move": &imaging.Move{}, "GetMoveOptions": &imaging.GetMoveOptions{}, "Stop": &imaging.Stop{}, "GetStatus": &imaging.GetStatus{}, "GetPresets": &imaging.GetPresets{}, "GetCurrentPreset": &imaging.GetCurrentPreset{}, "SetCurrentPreset": &imaging.SetCurrentPreset{}},
	"analytics": {"GetSupportedRules": &analytics.GetSupportedRules{}, "CreateRules": &analytics.CreateRules{}, "DeleteRules": &analytics.DeleteRules{}, "GetRules": &analytics.GetRules{}, "GetRuleOptions": &analytics.GetRuleOptions{}, "ModifyRules": &analytics.ModifyRules{}, "GetServiceCapabilities": &analytics.GetServiceCapabilities{}, "GetSupportedAnalyticsModules": &analytics.GetSupportedAnalyticsModules{}, "GetAnalyticsModuleOptions": &analytics.GetAnalyticsModuleOptions{}, "CreateAnalyticsModules": &analytics.CreateAnalyticsModules{}, "DeleteAnalyticsModules": &analytics.DeleteAnalyticsModules{}, "GetAnalyticsModules": &analytics.GetAnalyticsModules{}, "ModifyAnalyticsModules": &analytics.ModifyAnalyticsModules{}},
	"device":    {"Service": &device.Service{}, "Capabilities": &device.Capabilities{}, "DeviceServiceCapabilities": &device.DeviceServiceCapabilities{}, "NetworkCapabilities": &device.NetworkCapabilities{}, "SecurityCapabilities": &device.SecurityCapabilities{}, "EAPMethodTypes": &device.EAPMethodTypes{}, "SystemCapabilities": &device.SystemCapabilities{}, "MiscCapabilities": &device.MiscCapabilities{}, "StorageConfiguration": &device.StorageConfiguration{}, "StorageConfigurationData": &device.StorageConfigurationData{}, "UserCredential": &device.UserCredential{}, "GetServices": &device.GetServices{}, "GetServicesResponse": &device.GetServicesResponse{}, "GetServiceCapabilities": &device.GetServiceCapabilities{}, "GetServiceCapabilitiesResponse": &device.GetServiceCapabilitiesResponse{}, "GetDeviceInformation": &device.GetDeviceInformation{}, "GetDeviceInformationResponse": &device.GetDeviceInformationResponse{}, "SetSystemDateAndTime": &device.SetSystemDateAndTime{}, "SetSystemDateAndTimeResponse": &device.SetSystemDateAndTimeResponse{}, "GetSystemDateAndTime": &device.GetSystemDateAndTime{}, "GetSystemDateAndTimeResponse": &device.GetSystemDateAndTimeResponse{}, "SetSystemFactoryDefault": &device.SetSystemFactoryDefault{}, "SetSystemFactoryDefaultResponse": &device.SetSystemFactoryDefaultResponse{}, "UpgradeSystemFirmware": &device.UpgradeSystemFirmware{}, "UpgradeSystemFirmwareResponse": &device.UpgradeSystemFirmwareResponse{}, "SystemReboot": &device.SystemReboot{}, "SystemRebootResponse": &device.SystemRebootResponse{}, "RestoreSystem": &device.RestoreSystem{}, "RestoreSystemResponse": &device.RestoreSystemResponse{}, "GetSystemBackup": &device.GetSystemBackup{}, "GetSystemBackupResponse": &device.GetSystemBackupResponse{}, "GetSystemLog": &device.GetSystemLog{}, "GetSystemLogResponse": &device.GetSystemLogResponse{}, "GetSystemSupportInformation": &device.GetSystemSupportInformation{}, "GetSystemSupportInformationResponse": &device.GetSystemSupportInformationResponse{}, "GetScopes": &device.GetScopes{}, "GetScopesResponse": &device.GetScopesResponse{}, "SetScopes": &device.SetScopes{}, "SetScopesResponse": &device.SetScopesResponse{}, "AddScopes": &device.AddScopes{}, "AddScopesResponse": &device.AddScopesResponse{}, "RemoveScopes": &device.RemoveScopes{}, "RemoveScopesResponse": &device.RemoveScopesResponse{}, "GetDiscoveryMode": &device.GetDiscoveryMode{}, "GetDiscoveryModeResponse": &device.GetDiscoveryModeResponse{}, "SetDiscoveryMode": &device.SetDiscoveryMode{}, "SetDiscoveryModeResponse": &device.SetDiscoveryModeResponse{}, "GetRemoteDiscoveryMode": &device.GetRemoteDiscoveryMode{}, "GetRemoteDiscoveryModeResponse": &device.GetRemoteDiscoveryModeResponse{}, "SetRemoteDiscoveryMode": &device.SetRemoteDiscoveryMode{}, "SetRemoteDiscoveryModeResponse": &device.SetRemoteDiscoveryModeResponse{}, "GetDPAddresses": &device.GetDPAddresses{}, "GetDPAddressesResponse": &device.GetDPAddressesResponse{}, "SetDPAddresses": &device.SetDPAddresses{}, "SetDPAddressesResponse": &device.SetDPAddressesResponse{}, "GetEndpointReference": &device.GetEndpointReference{}, "GetEndpointReferenceResponse": &device.GetEndpointReferenceResponse{}, "GetRemoteUser": &device.GetRemoteUser{}, "GetRemoteUserResponse": &device.GetRemoteUserResponse{}, "SetRemoteUser": &device.SetRemoteUser{}, "SetRemoteUserResponse": &device.SetRemoteUserResponse{}, "GetUsers": &device.GetUsers{}, "GetUsersResponse": &device.GetUsersResponse{}, "CreateUsers": &device.CreateUsers{}, "CreateUsersResponse": &device.CreateUsersResponse{}, "DeleteUsers": &device.DeleteUsers{}, "DeleteUsersResponse": &device.DeleteUsersResponse{}, "SetUser": &device.SetUser{}, "SetUserResponse": &device.SetUserResponse{}, "GetWsdlUrl": &device.GetWsdlUrl{}, "GetWsdlUrlResponse": &device.GetWsdlUrlResponse{}, "GetCapabilities": &device.GetCapabilities{}, "GetCapabilitiesResponse": &device.GetCapabilitiesResponse{}, "GetHostname": &device.GetHostname{}, "GetHostnameResponse": &device.GetHostnameResponse{}, "SetHostname": &device.SetHostname{}, "SetHostnameResponse": &device.SetHostnameResponse{}, "SetHostnameFromDHCP": &device.SetHostnameFromDHCP{}, "SetHostnameFromDHCPResponse": &device.SetHostnameFromDHCPResponse{}, "GetDNS": &device.GetDNS{}, "GetDNSResponse": &device.GetDNSResponse{}, "SetDNS": &device.SetDNS{}, "SetDNSResponse": &device.SetDNSResponse{}, "GetNTP": &device.GetNTP{}, "GetNTPResponse": &device.GetNTPResponse{}, "SetNTP": &device.SetNTP{}, "SetNTPResponse": &device.SetNTPResponse{}, "GetDynamicDNS": &device.GetDynamicDNS{}, "GetDynamicDNSResponse": &device.GetDynamicDNSResponse{}, "SetDynamicDNS": &device.SetDynamicDNS{}, "SetDynamicDNSResponse": &device.SetDynamicDNSResponse{}, "GetNetworkInterfaces": &device.GetNetworkInterfaces{}, "GetNetworkInterfacesResponse": &device.GetNetworkInterfacesResponse{}, "SetNetworkInterfaces": &device.SetNetworkInterfaces{}, "SetNetworkInterfacesResponse": &device.SetNetworkInterfacesResponse{}, "GetNetworkProtocols": &device.GetNetworkProtocols{}, "GetNetworkProtocolsResponse": &device.GetNetworkProtocolsResponse{}, "SetNetworkProtocols": &device.SetNetworkProtocols{}, "SetNetworkProtocolsResponse": &device.SetNetworkProtocolsResponse{}, "GetNetworkDefaultGateway": &device.GetNetworkDefaultGateway{}, "GetNetworkDefaultGatewayResponse": &device.GetNetworkDefaultGatewayResponse{}, "SetNetworkDefaultGateway": &device.SetNetworkDefaultGateway{}, "SetNetworkDefaultGatewayResponse": &device.SetNetworkDefaultGatewayResponse{}, "GetZeroConfiguration": &device.GetZeroConfiguration{}, "GetZeroConfigurationResponse": &device.GetZeroConfigurationResponse{}, "SetZeroConfiguration": &device.SetZeroConfiguration{}, "SetZeroConfigurationResponse": &device.SetZeroConfigurationResponse{}, "GetIPAddressFilter": &device.GetIPAddressFilter{}, "GetIPAddressFilterResponse": &device.GetIPAddressFilterResponse{}, "SetIPAddressFilter": &device.SetIPAddressFilter{}, "SetIPAddressFilterResponse": &device.SetIPAddressFilterResponse{}, "AddIPAddressFilter": &device.AddIPAddressFilter{}, "AddIPAddressFilterResponse": &device.AddIPAddressFilterResponse{}, "RemoveIPAddressFilter": &device.RemoveIPAddressFilter{}, "RemoveIPAddressFilterResponse": &device.RemoveIPAddressFilterResponse{}, "GetAccessPolicy": &device.GetAccessPolicy{}, "GetAccessPolicyResponse": &device.GetAccessPolicyResponse{}, "SetAccessPolicy": &device.SetAccessPolicy{}, "SetAccessPolicyResponse": &device.SetAccessPolicyResponse{}, "CreateCertificate": &device.CreateCertificate{}, "CreateCertificateResponse": &device.CreateCertificateResponse{}, "GetCertificates": &device.GetCertificates{}, "GetCertificatesResponse": &device.GetCertificatesResponse{}, "GetCertificatesStatus": &device.GetCertificatesStatus{}, "GetCertificatesStatusResponse": &device.GetCertificatesStatusResponse{}, "SetCertificatesStatus": &device.SetCertificatesStatus{}, "SetCertificatesStatusResponse": &device.SetCertificatesStatusResponse{}, "DeleteCertificates": &device.DeleteCertificates{}, "DeleteCertificatesResponse": &device.DeleteCertificatesResponse{}, "GetPkcs10Request": &device.GetPkcs10Request{}, "GetPkcs10RequestResponse": &device.GetPkcs10RequestResponse{}, "LoadCertificates": &device.LoadCertificates{}, "LoadCertificatesResponse": &device.LoadCertificatesResponse{}, "GetClientCertificateMode": &device.GetClientCertificateMode{}, "GetClientCertificateModeResponse": &device.GetClientCertificateModeResponse{}, "SetClientCertificateMode": &device.SetClientCertificateMode{}, "SetClientCertificateModeResponse": &device.SetClientCertificateModeResponse{}, "GetRelayOutputs": &device.GetRelayOutputs{}, "GetRelayOutputsResponse": &device.GetRelayOutputsResponse{}, "SetRelayOutputSettings": &device.SetRelayOutputSettings{}, "SetRelayOutputSettingsResponse": &device.SetRelayOutputSettingsResponse{}, "SetRelayOutputState": &device.SetRelayOutputState{}, "SetRelayOutputStateResponse": &device.SetRelayOutputStateResponse{}, "SendAuxiliaryCommand": &device.SendAuxiliaryCommand{}, "SendAuxiliaryCommandResponse": &device.SendAuxiliaryCommandResponse{}, "GetCACertificates": &device.GetCACertificates{}, "GetCACertificatesResponse": &device.GetCACertificatesResponse{}, "LoadCertificateWithPrivateKey": &device.LoadCertificateWithPrivateKey{}, "LoadCertificateWithPrivateKeyResponse": &device.LoadCertificateWithPrivateKeyResponse{}, "GetCertificateInformation": &device.GetCertificateInformation{}, "GetCertificateInformationResponse": &device.GetCertificateInformationResponse{}, "LoadCACertificates": &device.LoadCACertificates{}, "LoadCACertificatesResponse": &device.LoadCACertificatesResponse{}, "CreateDot1XConfiguration": &device.CreateDot1XConfiguration{}, "CreateDot1XConfigurationResponse": &device.CreateDot1XConfigurationResponse{}, "SetDot1XConfiguration": &device.SetDot1XConfiguration{}, "SetDot1XConfigurationResponse": &device.SetDot1XConfigurationResponse{}, "GetDot1XConfiguration": &device.GetDot1XConfiguration{}, "GetDot1XConfigurationResponse": &device.GetDot1XConfigurationResponse{}, "GetDot1XConfigurations": &device.GetDot1XConfigurations{}, "GetDot1XConfigurationsResponse": &device.GetDot1XConfigurationsResponse{}, "DeleteDot1XConfiguration": &device.DeleteDot1XConfiguration{}, "DeleteDot1XConfigurationResponse": &device.DeleteDot1XConfigurationResponse{}, "GetDot11Capabilities": &device.GetDot11Capabilities{}, "GetDot11CapabilitiesResponse": &device.GetDot11CapabilitiesResponse{}, "GetDot11Status": &device.GetDot11Status{}, "GetDot11StatusResponse": &device.GetDot11StatusResponse{}, "ScanAvailableDot11Networks": &device.ScanAvailableDot11Networks{}, "ScanAvailableDot11NetworksResponse": &device.ScanAvailableDot11NetworksResponse{}, "GetSystemUris": &device.GetSystemUris{}, "GetSystemUrisResponse": &device.GetSystemUrisResponse{}, "StartFirmwareUpgrade": &device.StartFirmwareUpgrade{}, "StartFirmwareUpgradeResponse": &device.StartFirmwareUpgradeResponse{}, "StartSystemRestore": &device.StartSystemRestore{}, "StartSystemRestoreResponse": &device.StartSystemRestoreResponse{}, "GetStorageConfigurations": &device.GetStorageConfigurations{}, "GetStorageConfigurationsResponse": &device.GetStorageConfigurationsResponse{}, "CreateStorageConfiguration": &device.CreateStorageConfiguration{}, "CreateStorageConfigurationResponse": &device.CreateStorageConfigurationResponse{}, "GetStorageConfiguration": &device.GetStorageConfiguration{}, "GetStorageConfigurationResponse": &device.GetStorageConfigurationResponse{}, "SetStorageConfiguration": &device.SetStorageConfiguration{}, "SetStorageConfigurationResponse": &device.SetStorageConfigurationResponse{}, "DeleteStorageConfiguration": &device.DeleteStorageConfiguration{}, "DeleteStorageConfigurationResponse": &device.DeleteStorageConfigurationResponse{}, "GetGeoLocation": &device.GetGeoLocation{}, "GetGeoLocationResponse": &device.GetGeoLocationResponse{}, "SetGeoLocation": &device.SetGeoLocation{}, "SetGeoLocationResponse": &device.SetGeoLocationResponse{}, "DeleteGeoLocation": &device.DeleteGeoLocation{}, "DeleteGeoLocationResponse": &device.DeleteGeoLocationResponse{}},
	"event":     {"AbsoluteOrRelativeTimeType": &event.AbsoluteOrRelativeTimeType{}, "EndpointReferenceType": &event.EndpointReferenceType{}, "FilterType": &event.FilterType{}, "ReferenceParametersType": &event.ReferenceParametersType{}, "MetadataType": &event.MetadataType{}, "TopicSetType": &event.TopicSetType{}, "ExtensibleDocumented": &event.ExtensibleDocumented{}, "NotificationMessageHolderType": &event.NotificationMessageHolderType{}, "QueryExpressionType": &event.QueryExpressionType{}, "TopicExpressionType": &event.TopicExpressionType{}, "Capabilities": &event.Capabilities{}, "ResourceUnknownFault": &event.ResourceUnknownFault{}, "InvalidFilterFault": &event.InvalidFilterFault{}, "TopicExpressionDialectUnknownFault": &event.TopicExpressionDialectUnknownFault{}, "InvalidTopicExpressionFault": &event.InvalidTopicExpressionFault{}, "TopicNotSupportedFault": &event.TopicNotSupportedFault{}, "InvalidProducerPropertiesExpressionFault": &event.InvalidProducerPropertiesExpressionFault{}, "InvalidMessageContentExpressionFault": &event.InvalidMessageContentExpressionFault{}, "UnacceptableInitialTerminationTimeFault": &event.UnacceptableInitialTerminationTimeFault{}, "UnrecognizedPolicyRequestFault": &event.UnrecognizedPolicyRequestFault{}, "UnsupportedPolicyRequestFault": &event.UnsupportedPolicyRequestFault{}, "NotifyMessageNotSupportedFault": &event.NotifyMessageNotSupportedFault{}, "SubscribeCreationFailedFault": &event.SubscribeCreationFailedFault{}},
	"media":     {"Capabilities": &media.Capabilities{}, "ProfileCapabilities": &media.ProfileCapabilities{}, "StreamingCapabilities": &media.StreamingCapabilities{}, "GetServiceCapabilities": &media.GetServiceCapabilities{}, "GetServiceCapabilitiesResponse": &media.GetServiceCapabilitiesResponse{}, "GetVideoSources": &media.GetVideoSources{}, "GetVideoSourcesResponse": &media.GetVideoSourcesResponse{}, "GetAudioSources": &media.GetAudioSources{}, "GetAudioSourcesResponse": &media.GetAudioSourcesResponse{}, "GetAudioOutputs": &media.GetAudioOutputs{}, "GetAudioOutputsResponse": &media.GetAudioOutputsResponse{}, "CreateProfile": &media.CreateProfile{}, "CreateProfileResponse": &media.CreateProfileResponse{}, "GetProfile": &media.GetProfile{}, "GetProfileResponse": &media.GetProfileResponse{}, "GetProfiles": &media.GetProfiles{}, "GetProfilesResponse": &media.GetProfilesResponse{}, "AddVideoEncoderConfiguration": &media.AddVideoEncoderConfiguration{}, "AddVideoEncoderConfigurationResponse": &media.AddVideoEncoderConfigurationResponse{}, "RemoveVideoEncoderConfiguration": &media.RemoveVideoEncoderConfiguration{}, "RemoveVideoEncoderConfigurationResponse": &media.RemoveVideoEncoderConfigurationResponse{}, "AddVideoSourceConfiguration": &media.AddVideoSourceConfiguration{}, "AddVideoSourceConfigurationResponse": &media.AddVideoSourceConfigurationResponse{}, "RemoveVideoSourceConfiguration": &media.RemoveVideoSourceConfiguration{}, "RemoveVideoSourceConfigurationResponse": &media.RemoveVideoSourceConfigurationResponse{}, "AddAudioEncoderConfiguration": &media.AddAudioEncoderConfiguration{}, "AddAudioEncoderConfigurationResponse": &media.AddAudioEncoderConfigurationResponse{}, "RemoveAudioEncoderConfiguration": &media.RemoveAudioEncoderConfiguration{}, "RemoveAudioEncoderConfigurationResponse": &media.RemoveAudioEncoderConfigurationResponse{}, "AddAudioSourceConfiguration": &media.AddAudioSourceConfiguration{}, "AddAudioSourceConfigurationResponse": &media.AddAudioSourceConfigurationResponse{}, "RemoveAudioSourceConfiguration": &media.RemoveAudioSourceConfiguration{}, "RemoveAudioSourceConfigurationResponse": &media.RemoveAudioSourceConfigurationResponse{}, "AddPTZConfiguration": &media.AddPTZConfiguration{}, "AddPTZConfigurationResponse": &media.AddPTZConfigurationResponse{}, "RemovePTZConfiguration": &media.RemovePTZConfiguration{}, "RemovePTZConfigurationResponse": &media.RemovePTZConfigurationResponse{}, "AddVideoAnalyticsConfiguration": &media.AddVideoAnalyticsConfiguration{}, "AddVideoAnalyticsConfigurationResponse": &media.AddVideoAnalyticsConfigurationResponse{}, "RemoveVideoAnalyticsConfiguration": &media.RemoveVideoAnalyticsConfiguration{}, "RemoveVideoAnalyticsConfigurationResponse": &media.RemoveVideoAnalyticsConfigurationResponse{}, "AddMetadataConfiguration": &media.AddMetadataConfiguration{}, "AddMetadataConfigurationResponse": &media.AddMetadataConfigurationResponse{}, "RemoveMetadataConfiguration": &media.RemoveMetadataConfiguration{}, "RemoveMetadataConfigurationResponse": &media.RemoveMetadataConfigurationResponse{}, "AddAudioOutputConfiguration": &media.AddAudioOutputConfiguration{}, "AddAudioOutputConfigurationResponse": &media.AddAudioOutputConfigurationResponse{}, "RemoveAudioOutputConfiguration": &media.RemoveAudioOutputConfiguration{}, "RemoveAudioOutputConfigurationResponse": &media.RemoveAudioOutputConfigurationResponse{}, "AddAudioDecoderConfiguration": &media.AddAudioDecoderConfiguration{}, "AddAudioDecoderConfigurationResponse": &media.AddAudioDecoderConfigurationResponse{}, "RemoveAudioDecoderConfiguration": &media.RemoveAudioDecoderConfiguration{}, "RemoveAudioDecoderConfigurationResponse": &media.RemoveAudioDecoderConfigurationResponse{}, "DeleteProfile": &media.DeleteProfile{}, "DeleteProfileResponse": &media.DeleteProfileResponse{}, "GetVideoSourceConfigurations": &media.GetVideoSourceConfigurations{}, "GetVideoSourceConfigurationsResponse": &media.GetVideoSourceConfigurationsResponse{}, "GetVideoEncoderConfigurations": &media.GetVideoEncoderConfigurations{}, "GetVideoEncoderConfigurationsResponse": &media.GetVideoEncoderConfigurationsResponse{}, "GetAudioSourceConfigurations": &media.GetAudioSourceConfigurations{}, "GetAudioSourceConfigurationsResponse": &media.GetAudioSourceConfigurationsResponse{}, "GetAudioEncoderConfigurations": &media.GetAudioEncoderConfigurations{}, "GetAudioEncoderConfigurationsResponse": &media.GetAudioEncoderConfigurationsResponse{}, "GetVideoAnalyticsConfigurations": &media.GetVideoAnalyticsConfigurations{}, "GetVideoAnalyticsConfigurationsResponse": &media.GetVideoAnalyticsConfigurationsResponse{}, "GetMetadataConfigurations": &media.GetMetadataConfigurations{}, "GetMetadataConfigurationsResponse": &media.GetMetadataConfigurationsResponse{}, "GetAudioOutputConfigurations": &media.GetAudioOutputConfigurations{}, "GetAudioOutputConfigurationsResponse": &media.GetAudioOutputConfigurationsResponse{}, "GetAudioDecoderConfigurations": &media.GetAudioDecoderConfigurations{}, "GetAudioDecoderConfigurationsResponse": &media.GetAudioDecoderConfigurationsResponse{}, "GetVideoSourceConfiguration": &media.GetVideoSourceConfiguration{}, "GetVideoSourceConfigurationResponse": &media.GetVideoSourceConfigurationResponse{}, "GetVideoEncoderConfiguration": &media.GetVideoEncoderConfiguration{}, "GetVideoEncoderConfigurationResponse": &media.GetVideoEncoderConfigurationResponse{}, "GetAudioSourceConfiguration": &media.GetAudioSourceConfiguration{}, "GetAudioSourceConfigurationResponse": &media.GetAudioSourceConfigurationResponse{}, "GetAudioEncoderConfiguration": &media.GetAudioEncoderConfiguration{}, "GetAudioEncoderConfigurationResponse": &media.GetAudioEncoderConfigurationResponse{}, "GetVideoAnalyticsConfiguration": &media.GetVideoAnalyticsConfiguration{}, "GetVideoAnalyticsConfigurationResponse": &media.GetVideoAnalyticsConfigurationResponse{}, "GetMetadataConfiguration": &media.GetMetadataConfiguration{}, "GetMetadataConfigurationResponse": &media.GetMetadataConfigurationResponse{}, "GetAudioOutputConfiguration": &media.GetAudioOutputConfiguration{}, "GetAudioOutputConfigurationResponse": &media.GetAudioOutputConfigurationResponse{}, "GetAudioDecoderConfiguration": &media.GetAudioDecoderConfiguration{}, "GetAudioDecoderConfigurationResponse": &media.GetAudioDecoderConfigurationResponse{}, "GetCompatibleVideoEncoderConfigurations": &media.GetCompatibleVideoEncoderConfigurations{}, "GetCompatibleVideoEncoderConfigurationsResponse": &media.GetCompatibleVideoEncoderConfigurationsResponse{}, "GetCompatibleVideoSourceConfigurations": &media.GetCompatibleVideoSourceConfigurations{}, "GetCompatibleVideoSourceConfigurationsResponse": &media.GetCompatibleVideoSourceConfigurationsResponse{}, "GetCompatibleAudioEncoderConfigurations": &media.GetCompatibleAudioEncoderConfigurations{}, "GetCompatibleAudioEncoderConfigurationsResponse": &media.GetCompatibleAudioEncoderConfigurationsResponse{}, "GetCompatibleAudioSourceConfigurations": &media.GetCompatibleAudioSourceConfigurations{}, "GetCompatibleAudioSourceConfigurationsResponse": &media.GetCompatibleAudioSourceConfigurationsResponse{}, "GetCompatibleVideoAnalyticsConfigurations": &media.GetCompatibleVideoAnalyticsConfigurations{}, "GetCompatibleVideoAnalyticsConfigurationsResponse": &media.GetCompatibleVideoAnalyticsConfigurationsResponse{}, "GetCompatibleMetadataConfigurations": &media.GetCompatibleMetadataConfigurations{}, "GetCompatibleMetadataConfigurationsResponse": &media.GetCompatibleMetadataConfigurationsResponse{}, "GetCompatibleAudioOutputConfigurations": &media.GetCompatibleAudioOutputConfigurations{}, "GetCompatibleAudioOutputConfigurationsResponse": &media.GetCompatibleAudioOutputConfigurationsResponse{}, "GetCompatibleAudioDecoderConfigurations": &media.GetCompatibleAudioDecoderConfigurations{}, "GetCompatibleAudioDecoderConfigurationsResponse": &media.GetCompatibleAudioDecoderConfigurationsResponse{}, "SetVideoSourceConfiguration": &media.SetVideoSourceConfiguration{}, "SetVideoSourceConfigurationResponse": &media.SetVideoSourceConfigurationResponse{}, "SetVideoEncoderConfiguration": &media.SetVideoEncoderConfiguration{}, "SetVideoEncoderConfigurationResponse": &media.SetVideoEncoderConfigurationResponse{}, "SetAudioSourceConfiguration": &media.SetAudioSourceConfiguration{}, "SetAudioSourceConfigurationResponse": &media.SetAudioSourceConfigurationResponse{}, "SetAudioEncoderConfiguration": &media.SetAudioEncoderConfiguration{}, "SetAudioEncoderConfigurationResponse": &media.SetAudioEncoderConfigurationResponse{}, "SetVideoAnalyticsConfiguration": &media.SetVideoAnalyticsConfiguration{}, "SetVideoAnalyticsConfigurationResponse": &media.SetVideoAnalyticsConfigurationResponse{}, "SetMetadataConfiguration": &media.SetMetadataConfiguration{}, "SetMetadataConfigurationResponse": &media.SetMetadataConfigurationResponse{}, "SetAudioOutputConfiguration": &media.SetAudioOutputConfiguration{}, "SetAudioOutputConfigurationResponse": &media.SetAudioOutputConfigurationResponse{}, "SetAudioDecoderConfiguration": &media.SetAudioDecoderConfiguration{}, "SetAudioDecoderConfigurationResponse": &media.SetAudioDecoderConfigurationResponse{}, "GetVideoSourceConfigurationOptions": &media.GetVideoSourceConfigurationOptions{}, "GetVideoSourceConfigurationOptionsResponse": &media.GetVideoSourceConfigurationOptionsResponse{}, "GetVideoEncoderConfigurationOptions": &media.GetVideoEncoderConfigurationOptions{}, "GetVideoEncoderConfigurationOptionsResponse": &media.GetVideoEncoderConfigurationOptionsResponse{}, "GetAudioSourceConfigurationOptions": &media.GetAudioSourceConfigurationOptions{}, "GetAudioSourceConfigurationOptionsResponse": &media.GetAudioSourceConfigurationOptionsResponse{}, "GetAudioEncoderConfigurationOptions": &media.GetAudioEncoderConfigurationOptions{}, "GetAudioEncoderConfigurationOptionsResponse": &media.GetAudioEncoderConfigurationOptionsResponse{}, "GetMetadataConfigurationOptions": &media.GetMetadataConfigurationOptions{}, "GetMetadataConfigurationOptionsResponse": &media.GetMetadataConfigurationOptionsResponse{}, "GetAudioOutputConfigurationOptions": &media.GetAudioOutputConfigurationOptions{}, "GetAudioOutputConfigurationOptionsResponse": &media.GetAudioOutputConfigurationOptionsResponse{}, "GetAudioDecoderConfigurationOptions": &media.GetAudioDecoderConfigurationOptions{}, "GetAudioDecoderConfigurationOptionsResponse": &media.GetAudioDecoderConfigurationOptionsResponse{}, "GetGuaranteedNumberOfVideoEncoderInstances": &media.GetGuaranteedNumberOfVideoEncoderInstances{}, "GetGuaranteedNumberOfVideoEncoderInstancesResponse": &media.GetGuaranteedNumberOfVideoEncoderInstancesResponse{}, "GetStreamUri": &media.GetStreamUri{}, "GetStreamUriResponse": &media.GetStreamUriResponse{}, "StartMulticastStreaming": &media.StartMulticastStreaming{}, "StartMulticastStreamingResponse": &media.StartMulticastStreamingResponse{}, "StopMulticastStreaming": &media.StopMulticastStreaming{}, "StopMulticastStreamingResponse": &media.StopMulticastStreamingResponse{}, "SetSynchronizationPoint": &media.SetSynchronizationPoint{}, "SetSynchronizationPointResponse": &media.SetSynchronizationPointResponse{}, "GetSnapshotUri": &media.GetSnapshotUri{}, "GetSnapshotUriResponse": &media.GetSnapshotUriResponse{}, "GetVideoSourceModes": &media.GetVideoSourceModes{}, "GetVideoSourceModesResponse": &media.GetVideoSourceModesResponse{}, "SetVideoSourceMode": &media.SetVideoSourceMode{}, "SetVideoSourceModeResponse": &media.SetVideoSourceModeResponse{}, "GetOSDs": &media.GetOSDs{}, "GetOSDsResponse": &media.GetOSDsResponse{}, "GetOSD": &media.GetOSD{}, "GetOSDResponse": &media.GetOSDResponse{}, "GetOSDOptions": &media.GetOSDOptions{}, "GetOSDOptionsResponse": &media.GetOSDOptionsResponse{}, "SetOSD": &media.SetOSD{}, "SetOSDResponse": &media.SetOSDResponse{}, "CreateOSD": &media.CreateOSD{}, "CreateOSDResponse": &media.CreateOSDResponse{}, "DeleteOSD": &media.DeleteOSD{}, "DeleteOSDResponse": &media.DeleteOSDResponse{}},
	"ptz":       {"Capabilities": &ptz.Capabilities{}, "GetServiceCapabilities": &ptz.GetServiceCapabilities{}, "GetServiceCapabilitiesResponse": &ptz.GetServiceCapabilitiesResponse{}, "GetNodes": &ptz.GetNodes{}, "GetNodesResponse": &ptz.GetNodesResponse{}, "GetNode": &ptz.GetNode{}, "GetNodeResponse": &ptz.GetNodeResponse{}, "GetConfiguration": &ptz.GetConfiguration{}, "GetConfigurationResponse": &ptz.GetConfigurationResponse{}, "GetConfigurations": &ptz.GetConfigurations{}, "GetConfigurationsResponse": &ptz.GetConfigurationsResponse{}, "SetConfiguration": &ptz.SetConfiguration{}, "SetConfigurationResponse": &ptz.SetConfigurationResponse{}, "GetConfigurationOptions": &ptz.GetConfigurationOptions{}, "GetConfigurationOptionsResponse": &ptz.GetConfigurationOptionsResponse{}, "SendAuxiliaryCommand": &ptz.SendAuxiliaryCommand{}, "SendAuxiliaryCommandResponse": &ptz.SendAuxiliaryCommandResponse{}, "GetPresets": &ptz.GetPresets{}, "GetPresetsResponse": &ptz.GetPresetsResponse{}, "SetPreset": &ptz.SetPreset{}, "SetPresetResponse": &ptz.SetPresetResponse{}, "RemovePreset": &ptz.RemovePreset{}, "RemovePresetResponse": &ptz.RemovePresetResponse{}, "GotoPreset": &ptz.GotoPreset{}, "GotoPresetResponse": &ptz.GotoPresetResponse{}, "GotoHomePosition": &ptz.GotoHomePosition{}, "GotoHomePositionResponse": &ptz.GotoHomePositionResponse{}, "SetHomePosition": &ptz.SetHomePosition{}, "SetHomePositionResponse": &ptz.SetHomePositionResponse{}, "ContinuousMove": &ptz.ContinuousMove{}, "ContinuousMoveResponse": &ptz.ContinuousMoveResponse{}, "RelativeMove": &ptz.RelativeMove{}, "RelativeMoveResponse": &ptz.RelativeMoveResponse{}, "GetStatus": &ptz.GetStatus{}, "GetStatusResponse": &ptz.GetStatusResponse{}, "AbsoluteMove": &ptz.AbsoluteMove{}, "AbsoluteMoveResponse": &ptz.AbsoluteMoveResponse{}, "GeoMove": &ptz.GeoMove{}, "GeoMoveResponse": &ptz.GeoMoveResponse{}, "Stop": &ptz.Stop{}, "StopResponse": &ptz.StopResponse{}, "GetPresetTours": &ptz.GetPresetTours{}, "GetPresetToursResponse": &ptz.GetPresetToursResponse{}, "GetPresetTour": &ptz.GetPresetTour{}, "GetPresetTourResponse": &ptz.GetPresetTourResponse{}, "GetPresetTourOptions": &ptz.GetPresetTourOptions{}, "GetPresetTourOptionsResponse": &ptz.GetPresetTourOptionsResponse{}, "CreatePresetTour": &ptz.CreatePresetTour{}, "CreatePresetTourResponse": &ptz.CreatePresetTourResponse{}, "ModifyPresetTour": &ptz.ModifyPresetTour{}, "ModifyPresetTourResponse": &ptz.ModifyPresetTourResponse{}, "OperatePresetTour": &ptz.OperatePresetTour{}, "OperatePresetTourResponse": &ptz.OperatePresetTourResponse{}, "RemovePresetTour": &ptz.RemovePresetTour{}, "RemovePresetTourResponse": &ptz.RemovePresetTourResponse{}, "GetCompatibleConfigurations": &ptz.GetCompatibleConfigurations{}, "GetCompatibleConfigurationsResponse": &ptz.GetCompatibleConfigurationsResponse{}},
}

func GetOnvifStructByServiceAndMethod(servicename, methodname string) (interface{}, error) {
	if service, ok := GetOnvifStruct[servicename]; ok {
		if method, ok := service[methodname]; ok {
			if reflect.TypeOf(method).Kind() != reflect.Pointer {
				method = reflect.ValueOf(&method).Addr().Interface()
			}
			return method, nil
		} else {
			return nil, errors.New("There is no such method " + methodname + " in the " + servicename + " service")
		}
	} else {
		return nil, errors.New("There is no such service " + servicename)
	}
}

func GetXMLNode(xmlBody string, nodeName string) (*xml.Decoder, *xml.StartElement, error) {

	xmlBytes := bytes.NewBufferString(xmlBody)
	decodedXML := xml.NewDecoder(xmlBytes)

	for {
		token, err := decodedXML.Token()
		if err != nil {
			break
		}
		switch et := token.(type) {
		case xml.StartElement:
			if et.Name.Local == nodeName {
				return decodedXML, &et, nil
			}
		}
	}
	return nil, nil, fmt.Errorf("error in NodeName")
}

func CallOnvif(serviceName, methodName, acceptedData, username, password, xaddr string) (respone interface{}, err error) {
	var responestr string

	if responestr, err = CallNecessaryMethod(serviceName, methodName, acceptedData, username, password, xaddr); err != nil {
		return
	}

	responeName := methodName + "Response"
	if respone, err = GetOnvifStructByServiceAndMethod(serviceName, responeName); err != nil {
		return responestr, nil
	}
	// xmlAnalize(respone, &responestr)
	// var doc, node *xmlquery.Node
	// if doc, err = xmlquery.Parse(strings.NewReader(responestr)); err == nil {
	// 	if node, err = xmlquery.Query(doc, "//*[local-name() = '"+responeName+"']"); err == nil && node != nil {
	// 		responetmp := node.OutputXML(true)
	// 		err = xml.Unmarshal([]byte(responetmp), respone)
	// 	}
	// }
	if decodedXML, et, err := GetXMLNode(responestr, responeName); err == nil {
		if err = decodedXML.DecodeElement(respone, et); err == nil {
			return respone, nil
		}
	}
	return responestr, nil
}

func CallNecessaryMethodWithRetFilter(serviceName, methodName, acceptedData, username, password, xaddr, xpathfilter string) (map[string]string, string, error) {
	retMap := make(map[string]string)
	var err error
	var rettype = ""
	methdRespone, err := CallNecessaryMethod(serviceName, methodName, acceptedData, username, password, xaddr)
	if err != nil {
		return retMap, rettype, err
	}
	if len(xpathfilter) != 0 {
		if retMap, err := sutils.XmlStringFindElements(&methdRespone, xpathfilter); err == nil {
			return retMap, rettype, nil
		} else {
			return retMap, rettype, fmt.Errorf("%s: %s => %s", err.Error(), xpathfilter, methdRespone)
		}
	} else {
		if sutils.StringIsXml(&methdRespone) {
			rettype = "xml"
			retMap[rettype] = methdRespone
			return retMap, rettype, nil
		} else {
			return retMap, rettype, fmt.Errorf("respone is not xml: %s", methdRespone)
		}
	}
}

func GetEndpoint(service, xaddr string) (string, error) {
	dev, err := gonvif.NewDevice(gonvif.DeviceParams{Xaddr: xaddr})
	if err != nil {
		return "", err
	}

	pkg := strings.ToLower(service)
	//	log.Debug("GetServices:", dev.GetServices(), ":end")
	var endpoint string

	if endpoint = dev.GetEndpoint(pkg); len(endpoint) != 0 {
		return endpoint, nil
	} else {
		return "", errors.New("There is no such endpoint for " + pkg)
	}
}

func GetIPFromuuid(device_id string, camsDiscovery []snetutils.CamerasInfo) string {
	for _, v := range camsDiscovery {
		if v.UUID == device_id {
			return v.IP
		}
	}
	return ""
}

func GetCaminfoFromuuid(device_id string, camsDiscovery []snetutils.CamerasInfo) snetutils.CamerasInfo {
	for _, v := range camsDiscovery {
		if v.UUID == device_id {
			return v
		}
	}
	return snetutils.CamerasInfo{}
}

func GetXaddrFromuuid(device_id string, camsDiscovery []snetutils.CamerasInfo) string {
	for _, v := range camsDiscovery {
		if v.UUID == device_id {
			return v.XADDR
		}
	}
	return ""
}

func GetProfiles(xaddr, username, password string) (map[string]string, error) {
	var err error

	retstr, retype, err := CallNecessaryMethodWithRetFilter("media", "GetProfiles", "", username, password, xaddr, "//*[local-name() = 'Profiles']/@token")
	if err != nil {
		return retstr, err
	}

	if retype == "xml" {
		return retstr, errors.New("can not found xpath")
	}

	return retstr, nil
}

func GetSnapshortUrls(xaddr, username, password string) ([]string, error) {
	retstr := []string{}
	var err error

	profilestoken, retype, err := CallNecessaryMethodWithRetFilter("media", "GetProfiles", "", username, password, xaddr, "//*[local-name() = 'Profiles']/@token")
	if err != nil {
		return retstr, err
	}

	if retype == "xml" {
		return retstr, errors.New("can not found xpath")
	}

	for _, v := range profilestoken {
		//		log.Debug("profilestoken", v)
		soappara := "<onvif><trt:ProfileToken>" + v + "</trt:ProfileToken></onvif>"

		urlsnapshotr, retype, err := CallNecessaryMethodWithRetFilter("media", "GetSnapshotUri", soappara, username, password, xaddr, "//*[local-name() = 'MediaUri']/node()[local-name() = 'Uri']")
		if err != nil {
			return retstr, err
		}

		if retype == "xml" {
			return retstr, errors.New("can not found xpath")
		}

		for _, v1 := range urlsnapshotr {
			retstr = append(retstr, v1)
		}

	}
	sort.Strings(retstr)
	return retstr, nil
}

func IsStreamOnline(link string) (ok bool) {
	slogrus.Debug(link)
	if !func() (ok bool) {
		ok = false
		c := gortsplib.Client{}

		u, err := url.Parse(link)
		if err != nil {
			return
		}
		c.ReadTimeout = time.Second
		err = c.Start(u.Scheme, u.Host)
		if err != nil {
			return
		}
		defer c.Close()
		// c.Options(u)
		_, _, _, err = c.Describe(u)
		if err != nil {
			return
		}
		return true
	}() {
		return false
		// log.Printf("available tracks: %v\n", tracks)
		// snetutils.DialExpec(link, "rtsp", time.Microsecond*500)
		if _, _, err := sexec.ExecCommandShellTimeout(fmt.Sprintf("ffprobe -v quiet -print_format json -show_format '%s'", link), time.Second*10); err == nil {
			return true
		} else {
			return false
		}
	} else {
		return true
	}
}

//	func GetDeviceInformation(xaddrOrCamip, username, password string) (device.GetDeviceInformationResponse, error) {
//		CallNecessaryMethodWithRetFilter("device", "GetDeviceInformation", "", username, password, xaddrOrCamip, "")
//	}
type CamStream struct {
	Model        string   `json:"model,omitempty"`
	MacVendor    string   `json:"macvendor,omitempty"`
	Paths        []string `json:"paths"`
	Manufacturer string   `json:"manufacturer,omitempty"`
}

var CamStreamsCache = []CamStream{}
var camStreamLastCommit string

func GetStreamUrls(xaddrOrCamip, username, password string) (stream_urls []string, err error) {
	stream_urls = make([]string, 0)
	var profilestoken map[string]string
	profilestoken, err = GetProfiles(xaddrOrCamip, username, password)
	// profilestoken, retype, err := onvifutils.CallNecessaryMethodWithRetFilter("media", "GetProfiles", "", username, password, xaddrOrCamip, "//*[local-name() = 'Profiles']/@token")
	if err != nil {
		return
	}

	ipcam := sregexp.New(sutils.Ipv4Regex).FindString(xaddrOrCamip)
	if len(ipcam) == 0 {
		return stream_urls, fmt.Errorf("can not get stream link of %s", xaddrOrCamip)
	}
	for _, v := range profilestoken {
		//		log.Debug("profilestoken", v)
		soappara := "<onvif><trt:ProfileToken>" + v + "</trt:ProfileToken></onvif>"
		if stream_urls_map, retype, err := CallNecessaryMethodWithRetFilter("media", "GetStreamUri", soappara, username, password, xaddrOrCamip, "//*[local-name() = 'MediaUri']/node()[local-name() = 'Uri']"); err == nil && retype != "xml" {
			for _, stream_url := range stream_urls_map {
				stream_urls = append(stream_urls, stream_url)
			}
			break
		}

	}
	if len(stream_urls) == 0 {
		var resp *gcurl.Response
		var dev *gonvif.Device
		dev, err = gonvif.NewDevice(gonvif.DeviceParams{Xaddr: xaddrOrCamip, Username: username, Password: password})
		if err != nil {
			return
		}
		// curl -H "Accept: application/vnd.github.VERSION.sha"
		if resp, err = gcurl.GetDefaultRequest().WithHeader("Accept", "application/vnd.github.VERSION.sha").Get("https://api.github.com/repos/sonnt85/camstreamlist/commits/main"); err == nil {
			if currentCommit, err := resp.Text(); err == nil && camStreamLastCommit != currentCommit {
				if resp, err = gcurl.Get("https://raw.githubusercontent.com/sonnt85/camstreamlist/main/camstreamlist.json"); err == nil {
					streams := make([]CamStream, 0)
					if err = resp.JSONUnmarshal(&streams); err == nil && len(streams) != 0 {
						camStreamLastCommit = currentCommit
						CamStreamsCache = streams
					}
				}
			}
		}

		mac := ""
		for i := 0; i < len(CamStreamsCache); i++ {
			if len(CamStreamsCache[i].Model) != 0 {
				if sregexp.New(CamStreamsCache[i].Model).MatchString(dev.Model) {
					stream_urls = CamStreamsCache[i].Paths
					break
				}
			}

			if len(CamStreamsCache[i].MacVendor) != 0 {
				if len(mac) == 0 {
					mac, _ = snetutils.MacFromIP(ipcam)
					mac = strings.ReplaceAll(mac, ":", "")
				}
				if len(mac) != 0 {
					if sregexp.New(CamStreamsCache[i].MacVendor).MatchString(mac) {
						stream_urls = CamStreamsCache[i].Paths
						break
					}
				}
			}
			if len(CamStreamsCache[i].Manufacturer) != 0 {
				if sregexp.New(CamStreamsCache[i].Manufacturer).MatchString(dev.Manufacturer) {
					stream_urls = CamStreamsCache[i].Paths
					break
				}
			}
		}
	}
	retstream_urls := make([]string, 0)
	if len(stream_urls) != 0 {
		for i := range stream_urls {
			stream_urls[i] = sregexp.New(sutils.Ipv4Regex).ReplaceAllString(stream_urls[i], ipcam)
			if len(username) != 0 || len(password) != 0 {
				stream_urls[i] = strings.Replace(stream_urls[i], "://", "://"+username+":"+password+"@", 1)
			}
			if IsStreamOnline(stream_urls[i]) {
				retstream_urls = append(retstream_urls, stream_urls[i])
				err = nil
				// return []string{stream_urls[i]}, nil
			}
		}
	}
	sort.Strings(retstream_urls)
	return retstream_urls, err
}

func CallNecessaryMethod(serviceName, methodName, acceptedData, username, password, xaddr string) (string, error) {
	var methodStruct interface{}
	var err error

	serviceName = strings.ToLower(serviceName)

	if methodStruct, err = GetOnvifStructByServiceAndMethod(serviceName, methodName); err != nil {
		return "", err
	}
	noparserXML := false

	if len(acceptedData) == 0 {
		//		soaptmp := gosoap.NewEmptySOAP()
		//		acceptedData = soaptmp.String()
		acceptedData = "<onvif></onvif>"
		//		acceptedData = "<" + methodName + ">" + "</" + methodName + ">"
		//		noparserXML = true
	} else {

		for k, v := range sregexp.New(`<onvif>((?s:.*))</onvif>`).FindStringSubmatch(acceptedData) {
			if k == 1 {
				methodTagName, err := xml.Marshal(methodStruct)
				if err != nil {
					return "", err
				}
				//				log.Println("acceptedData:", string(output), ":end\n", acceptedData)
				for k435, v435 := range sregexp.New(`^<([^>]+)>`).FindStringSubmatch(string(methodTagName)) {
					if k435 == 1 {
						//				acceptedData = `<onvif>"` + v + `"</onvif>`
						methodName = v435
						break
					}
				}

				acceptedData = "<" + methodName + ">" + v + "</" + methodName + ">"
				noparserXML = true
				break
			}
		}
		//		acceptedData = "<onvif>\"" + acceptedData + "\"</onvif>"
	}

	//	log.Println("acceptedData(converted):" + acceptedData + ":end)
	resp := ""
	if !noparserXML {
		respptr, err := xmlAnalize(methodStruct, &acceptedData) // todo: update to parser xml
		if err != nil {
			log.Error("xmlAnalize:", err)
			return "", err
		}

		resp = *respptr
	} else {
		resp = acceptedData
	}

	endpoint, err := GetEndpoint(serviceName, xaddr)
	if err != nil {
		return "", err
	}

	soap := gosoap.NewEmptySOAP()
	soap.AddStringBodyContent(resp)
	soap.AddRootNamespaces(gonvif.Xlmns)
	soap.AddWSSecurity(username, password)

	servResp, err := networking.SendSoapWithTimeout(new(http.Client), endpoint, []byte(soap.String()), time.Millisecond*3000)
	if err != nil {
		log.Error("SendSoap error:", err, ":end")
		return "", err
	}

	rsp, err := io.ReadAll(servResp.Body)
	if err != nil {
		return "", err
	}

	return string(rsp), nil
}

func xmlAnalize(methodStruct interface{}, acceptedData *string) (*string, error) {
	test := make([]map[string]string, 0)      //tags
	testunMarshal := make([][]interface{}, 0) //data
	var mas []string                          //idnt

	soapHandling(methodStruct, &test)
	test = mapProcessing(test)

	doc := etree.NewDocument()
	if err := doc.ReadFromString(*acceptedData); err != nil {
		return nil, err
	}
	etr := doc.FindElements("./*")
	xmlUnmarshal(etr, &testunMarshal, &mas)
	ident(&mas)

	document := etree.NewDocument()
	var el *etree.Element
	var idntIndex = 0

	for lstIndex := 0; lstIndex < len(testunMarshal); {
		lst := (testunMarshal)[lstIndex]
		elemName, attr, value, err := xmlMaker(&lst, &test, lstIndex)
		if err != nil {
			return nil, err
		}

		if mas[lstIndex] == "Push" && lstIndex == 0 { //done
			el = document.CreateElement(elemName)
			el.SetText(value)
			if len(attr) != 0 {
				for key, value := range attr {
					el.CreateAttr(key, value)
				}
			}
		} else if mas[idntIndex] == "Push" {
			pushTmp := etree.NewElement(elemName)
			pushTmp.SetText(value)
			if len(attr) != 0 {
				for key, value := range attr {
					pushTmp.CreateAttr(key, value)
				}
			}
			el.AddChild(pushTmp)
			el = pushTmp
		} else if mas[idntIndex] == "PushPop" {
			popTmp := etree.NewElement(elemName)
			popTmp.SetText(value)
			if len(attr) != 0 {
				for key, value := range attr {
					popTmp.CreateAttr(key, value)
				}
			}
			if el == nil {
				document.AddChild(popTmp)
			} else {
				el.AddChild(popTmp)
			}
		} else if mas[idntIndex] == "Pop" {
			el = el.Parent()
			lstIndex -= 1
		}
		idntIndex += 1
		lstIndex += 1
	}

	resp, err := document.WriteToString()
	if err != nil {
		return nil, err
	}

	return &resp, err
}

func xmlMaker(lst *[]interface{}, tags *[]map[string]string, lstIndex int) (string, map[string]string, string, error) {
	var elemName, value string
	attr := make(map[string]string)
	for tgIndx, tg := range *tags {
		if tgIndx == lstIndex {
			for index, elem := range *lst {
				if reflect.TypeOf(elem).String() == "[]etree.Attr" {
					conversion := elem.([]etree.Attr)
					for _, i := range conversion {
						attr[i.Key] = i.Value
					}
				} else {
					conversion := elem.(string)
					if index == 0 && lstIndex == 0 {
						res, err := xmlProcessing(tg["XMLName"])
						if err != nil {
							return "", nil, "", err
						}
						elemName = res
					} else if index == 0 {
						res, err := xmlProcessing(tg[conversion])
						if err != nil {
							return "", nil, "", err
						}
						elemName = res
					} else {
						value = conversion
					}
				}
			}
		}
	}
	return elemName, attr, value, nil
}

func xmlProcessing(tg string) (string, error) {
	r, _ := regexp.Compile(`"(.*?)"`)
	str := r.FindStringSubmatch(tg)
	if len(str) == 0 {
		return "", errors.New("out of range")
	}
	attr := strings.Index(str[1], ",attr")
	omit := strings.Index(str[1], ",omitempty")
	attrOmit := strings.Index(str[1], ",attr,omitempty")
	omitAttr := strings.Index(str[1], ",omitempty,attr")

	if attr > -1 && attrOmit == -1 && omitAttr == -1 {
		return str[1][0:attr], nil
	} else if omit > -1 && attrOmit == -1 && omitAttr == -1 {
		return str[1][0:omit], nil
	} else if attr == -1 && omit == -1 {
		return str[1], nil
	} else if attrOmit > -1 {
		return str[1][0:attrOmit], nil
	} else {
		return str[1][0:omitAttr], nil
	}

	// return "", errors.New("something went wrong")
}

func mapProcessing(mapVar []map[string]string) []map[string]string {
	for indx := 0; indx < len(mapVar); indx++ {
		element := mapVar[indx]
		for _, value := range element {
			if value == "" {
				mapVar = append(mapVar[:indx], mapVar[indx+1:]...)
				indx--
			}
			if strings.Contains(value, ",attr") {
				mapVar = append(mapVar[:indx], mapVar[indx+1:]...)
				indx--
			}
		}
	}
	return mapVar
}

func soapHandling(tp interface{}, tags *[]map[string]string) {
	s := reflect.ValueOf(tp).Elem()
	typeOfT := s.Type()
	if s.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		tmp, err := typeOfT.FieldByName(typeOfT.Field(i).Name)
		if !err {
			log.Error(err)
		}
		*tags = append(*tags, map[string]string{typeOfT.Field(i).Name: string(tmp.Tag)})
		subStruct := reflect.New(reflect.TypeOf(f.Interface()))
		soapHandling(subStruct.Interface(), tags)
	}
}

func xmlUnmarshal(elems []*etree.Element, data *[][]interface{}, mas *[]string) {
	for _, elem := range elems {
		*data = append(*data, []interface{}{elem.Tag, elem.Attr, elem.Text()})
		*mas = append(*mas, "Push")
		xmlUnmarshal(elem.FindElements("./*"), data, mas)
		*mas = append(*mas, "Pop")
	}
}

func ident(mas *[]string) {
	var buffer string
	for _, j := range *mas {
		buffer += j + " "
	}
	buffer = strings.Replace(buffer, "Push Pop ", "PushPop ", -1)
	buffer = strings.TrimSpace(buffer)
	*mas = strings.Split(buffer, " ")
}
