/*
 *
 * Copyright 2014 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package credentials implements various credentials supported by gRPC library,
// which encapsulate all the state needed by a client to authenticate with a
// server and make various assertions, e.g., about the client's identity, role,
// or whether it is authorized to make a particular call.
package credentials // import "google.golang.org/grpc-forked/credentials"

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc-forked/internal"
)

// PerRPCCredentials defines the common interface for the credentials which need to
// attach security information to every RPC (e.g., oauth2).
type PerRPCCredentials interface {
	// GetRequestMetadata gets the current request metadata, refreshing
	// tokens if required. This should be called by the transport layer on
	// each request, and the data should be populated in headers or other
	// context. If a status code is returned, it will be used as the status
	// for the RPC. uri is the URI of the entry point for the request.
	// When supported by the underlying implementation, ctx can be used for
	// timeout and cancellation. Additionally, RequestInfo data will be
	// available via ctx to this call.
	// TODO(zhaoq): Define the set of the qualified keys instead of leaving
	// it as an arbitrary string.
	GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error)
	// RequireTransportSecurity indicates whether the credentials requires
	// transport security.
	RequireTransportSecurity() bool
}

// SecurityLevel defines the protection level on an established connection.
//
// This API is experimental.
type SecurityLevel int

const (
	// NoSecurity indicates a connection is insecure.
	// The zero SecurityLevel value is invalid for backward compatibility.
	NoSecurity SecurityLevel = iota + 1
	// IntegrityOnly indicates a connection only provides integrity protection.
	IntegrityOnly
	// PrivacyAndIntegrity indicates a connection provides both privacy and integrity protection.
	PrivacyAndIntegrity
)

// String returns SecurityLevel in a string format.
func (s SecurityLevel) String() string {
	switch s {
	case NoSecurity:
		return "NoSecurity"
	case IntegrityOnly:
		return "IntegrityOnly"
	case PrivacyAndIntegrity:
		return "PrivacyAndIntegrity"
	}
	return fmt.Sprintf("invalid SecurityLevel: %v", int(s))
}

// CommonAuthInfo contains authenticated information common to AuthInfo implementations.
// It should be embedded in a struct implementing AuthInfo to provide additional information
// about the credentials.
//
// This API is experimental.
type CommonAuthInfo struct {
	SecurityLevel SecurityLevel
}

// GetCommonAuthInfo returns the pointer to CommonAuthInfo struct.
func (c *CommonAuthInfo) GetCommonAuthInfo() *CommonAuthInfo {
	return c
}

// ProtocolInfo provides information regarding the gRPC wire protocol version,
// security protocol, security protocol version in use, server name, etc.
type ProtocolInfo struct {
	// ProtocolVersion is the gRPC wire protocol version.
	ProtocolVersion string
	// SecurityProtocol is the security protocol in use.
	SecurityProtocol string
	// SecurityVersion is the security protocol version.  It is a static version string from the
	// credentials, not a value that reflects per-connection protocol negotiation.  To retrieve
	// details about the credentials used for a connection, use the Peer's AuthInfo field instead.
	//
	// Deprecated: please use Peer.AuthInfo.
	SecurityVersion string
	// ServerName is the user-configured server name.
	ServerName string
}

// AuthInfo defines the common interface for the auth information the users are interested in.
// A struct that implements AuthInfo should embed CommonAuthInfo by including additional
// information about the credentials in it.
type AuthInfo interface {
	AuthType() string
}

// ErrConnDispatched indicates that rawConn has been dispatched out of gRPC
// and the caller should not close rawConn.
var ErrConnDispatched = errors.New("credentials: rawConn is dispatched out of gRPC")

// TransportCredentials defines the common interface for all the live gRPC wire
// protocols and supported transport security protocols (e.g., TLS, SSL).
type TransportCredentials interface {
	// ClientHandshake does the authentication handshake specified by the corresponding
	// authentication protocol on rawConn for clients. It returns the authenticated
	// connection and the corresponding auth information about the connection.
	// The auth information should embed CommonAuthInfo to return additional information about
	// the credentials. Implementations must use the provided context to implement timely cancellation.
	// gRPC will try to reconnect if the error returned is a temporary error
	// (io.EOF, context.DeadlineExceeded or err.Temporary() == true).
	// If the returned error is a wrapper error, implementations should make sure that
	// the error implements Temporary() to have the correct retry behaviors.
	//
	// If the returned net.Conn is closed, it MUST close the net.Conn provided.
	ClientHandshake(context.Context, string, net.Conn) (net.Conn, AuthInfo, error)
	// ServerHandshake does the authentication handshake for servers. It returns
	// the authenticated connection and the corresponding auth information about
	// the connection. The auth information should embed CommonAuthInfo to return additional information
	// about the credentials.
	//
	// If the returned net.Conn is closed, it MUST close the net.Conn provided.
	ServerHandshake(net.Conn) (net.Conn, AuthInfo, error)
	// Info provides the ProtocolInfo of this TransportCredentials.
	Info() ProtocolInfo
	// Clone makes a copy of this TransportCredentials.
	Clone() TransportCredentials
	// OverrideServerName overrides the server name used to verify the hostname on the returned certificates from the server.
	// gRPC internals also use it to override the virtual hosting name if it is set.
	// It must be called before dialing. Currently, this is only used by grpclb.
	OverrideServerName(string) error
}

// Bundle is a combination of TransportCredentials and PerRPCCredentials.
//
// It also contains a mode switching method, so it can be used as a combination
// of different credential policies.
//
// Bundle cannot be used together with individual TransportCredentials.
// PerRPCCredentials from Bundle will be appended to other PerRPCCredentials.
//
// This API is experimental.
type Bundle interface {
	TransportCredentials() TransportCredentials
	PerRPCCredentials() PerRPCCredentials
	// NewWithMode should make a copy of Bundle, and switch mode. Modifying the
	// existing Bundle may cause races.
	//
	// NewWithMode returns nil if the requested mode is not supported.
	NewWithMode(mode string) (Bundle, error)
}

// RequestInfo contains request data attached to the context passed to GetRequestMetadata calls.
//
// This API is experimental.
type RequestInfo struct {
	// The method passed to Invoke or NewStream for this RPC. (For proto methods, this has the format "/some.Service/Method")
	Method string
	// AuthInfo contains the information from a security handshake (TransportCredentials.ClientHandshake, TransportCredentials.ServerHandshake)
	AuthInfo AuthInfo
}

// requestInfoKey is a struct to be used as the key when attaching a RequestInfo to a context object.
type requestInfoKey struct{}

// RequestInfoFromContext extracts the RequestInfo from the context if it exists.
//
// This API is experimental.
func RequestInfoFromContext(ctx context.Context) (ri RequestInfo, ok bool) {
	ri, ok = ctx.Value(requestInfoKey{}).(RequestInfo)
	return
}

// CheckSecurityLevel checks if a connection's security level is greater than or equal to the specified one.
// It returns success if 1) the condition is satisified or 2) AuthInfo struct does not implement GetCommonAuthInfo() method
// or 3) CommonAuthInfo.SecurityLevel has an invalid zero value. For 2) and 3), it is for the purpose of backward-compatibility.
//
// This API is experimental.
func CheckSecurityLevel(ctx context.Context, level SecurityLevel) error {
	type internalInfo interface {
		GetCommonAuthInfo() *CommonAuthInfo
	}
	ri, _ := RequestInfoFromContext(ctx)
	if ri.AuthInfo == nil {
		return errors.New("unable to obtain SecurityLevel from context")
	}
	if ci, ok := ri.AuthInfo.(internalInfo); ok {
		// CommonAuthInfo.SecurityLevel has an invalid value.
		if ci.GetCommonAuthInfo().SecurityLevel == 0 {
			return nil
		}
		if ci.GetCommonAuthInfo().SecurityLevel < level {
			return fmt.Errorf("requires SecurityLevel %v; connection has %v", level, ci.GetCommonAuthInfo().SecurityLevel)
		}
	}
	// The condition is satisfied or AuthInfo struct does not implement GetCommonAuthInfo() method.
	return nil
}

func init() {
	internal.NewRequestInfoContext = func(ctx context.Context, ri RequestInfo) context.Context {
		return context.WithValue(ctx, requestInfoKey{}, ri)
	}
}

// ChannelzSecurityInfo defines the interface that security protocols should implement
// in order to provide security info to channelz.
//
// This API is experimental.
type ChannelzSecurityInfo interface {
	GetSecurityValue() ChannelzSecurityValue
}

// ChannelzSecurityValue defines the interface that GetSecurityValue() return value
// should satisfy. This interface should only be satisfied by *TLSChannelzSecurityValue
// and *OtherChannelzSecurityValue.
//
// This API is experimental.
type ChannelzSecurityValue interface {
	isChannelzSecurityValue()
}

// OtherChannelzSecurityValue defines the struct that non-TLS protocol should return
// from GetSecurityValue(), which contains protocol specific security info. Note
// the Value field will be sent to users of channelz requesting channel info, and
// thus sensitive info should better be avoided.
//
// This API is experimental.
type OtherChannelzSecurityValue struct {
	ChannelzSecurityValue
	Name  string
	Value proto.Message
}
