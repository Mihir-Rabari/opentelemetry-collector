// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package pprofile // import "go.opentelemetry.io/collector/pdata/pprofile"

import "go.opentelemetry.io/collector/pdata/internal/otlp"

var _ MarshalSizer = (*ProtoMarshaler)(nil)

type ProtoMarshaler struct{}

// MarshalProfiles marshals Profiles to gRPC format bytes.
// If the input data is read-only, it will be copied to a mutable
// instance before mutation.
func (e *ProtoMarshaler) MarshalProfiles(pd Profiles) ([]byte, error) {
	// Only copy if data is shared/read-only to avoid unnecessary allocation
	pdToUse := pd
	if pd.IsReadOnly() {
		pdCopy := NewProfiles()
		pd.CopyTo(pdCopy)
		pdToUse = pdCopy
	}

	// Convert strings to references for efficient transmission
	convertProfilesToReferences(pdToUse)

	size := pdToUse.getOrig().SizeProto()
	buf := make([]byte, size)
	_ = pdToUse.getOrig().MarshalProto(buf)
	return buf, nil
}

// ProfilesSize returns the size in bytes of a marshaled Profiles. This
// matches the length of the bytes MarshalProfiles produces: MarshalProfiles
// converts resource and scope attribute strings to dictionary references
// before marshaling, so the size is computed on a converted copy of the
// input, which is left unmodified.
func (e *ProtoMarshaler) ProfilesSize(pd Profiles) int {
	pdToSize := NewProfiles()
	pd.CopyTo(pdToSize)
	convertProfilesToReferences(pdToSize)
	return pdToSize.getOrig().SizeProto()
}

func (e *ProtoMarshaler) ResourceProfilesSize(pd ResourceProfiles) int {
	return pd.orig.SizeProto()
}

func (e *ProtoMarshaler) ScopeProfilesSize(pd ScopeProfiles) int {
	return pd.orig.SizeProto()
}

func (e *ProtoMarshaler) ProfileSize(pd Profile) int {
	return pd.orig.SizeProto()
}

type ProtoUnmarshaler struct{}

func (d *ProtoUnmarshaler) UnmarshalProfiles(buf []byte) (Profiles, error) {
	pd := NewProfiles()
	err := pd.getOrig().UnmarshalProto(buf)
	if err != nil {
		return Profiles{}, err
	}
	otlp.MigrateProfiles(pd.getOrig().ResourceProfiles)

	// Resolve all string_value_ref and key_ref to their actual strings
	// so the pdata API works transparently
	resolveProfilesReferences(pd)

	return pd, nil
}
