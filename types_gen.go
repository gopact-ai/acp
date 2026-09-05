// SPDX-FileCopyrightText: 2025 The Go MCP SDK Authors
// SPDX-FileCopyrightText: 2026 gopact-ai contributors
// SPDX-License-Identifier: Apache-2.0 AND MIT

// Generated with github.com/spachava753/acp-sdk/internal/schemagen at
// ea76600dde1bd490a2fc6c0c4a44f05383a8abc9 from stable ACP v1 at protocol
// commit af41b25f57a79c5629b3164e23fb4e8650badeeb.
// Modified by gopact-ai to enforce required union fields and preserve concrete
// Go types when decoding stable-v1 union variants.
// PromptResponse.Usage and Usage are maintained by hand from
// @agentclientprotocol/sdk v1.4.0 schema/schema.json because this repository
// has no generation script and regenerating would lose the corrections above.

package acp

import (
	"bytes"
	"encoding/json"
	"errors"
	"maps"
	"reflect"
)

// Meta: Reserved metadata for protocol extensions.
type Meta map[string]any

// AgentAuthCapabilities: Authentication-related capabilities supported by the agent.
type AgentAuthCapabilities struct {
	Meta   Meta                `json:"_meta,omitzero"`
	Logout *LogoutCapabilities `json:"logout,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *AgentAuthCapabilities) UnmarshalJSON(data []byte) error {
	type alias AgentAuthCapabilities
	decoded := alias{}
	raw := struct {
		Logout json.RawMessage `json:"logout"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Logout) > 0 {
		unmarshalDefault(raw.Logout, &decoded.Logout)
	}
	*c = AgentAuthCapabilities(decoded)
	return nil
}

// AgentCapabilities: Capabilities supported by the agent.
//
// Advertised during initialization to inform the client about
// available features and content types.
//
// See protocol docs: [Agent Capabilities](https://agentclientprotocol.com/protocol/initialization#agent-capabilities)
type AgentCapabilities struct {
	Meta                Meta                   `json:"_meta,omitzero"`
	Auth                *AgentAuthCapabilities `json:"auth,omitempty"`
	LoadSession         bool                   `json:"loadSession,omitempty"`
	MCPCapabilities     *MCPCapabilities       `json:"mcpCapabilities,omitempty"`
	PromptCapabilities  *PromptCapabilities    `json:"promptCapabilities,omitempty"`
	SessionCapabilities *SessionCapabilities   `json:"sessionCapabilities,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *AgentCapabilities) UnmarshalJSON(data []byte) error {
	type alias AgentCapabilities
	decoded := alias{}
	raw := struct {
		Auth                json.RawMessage `json:"auth"`
		LoadSession         json.RawMessage `json:"loadSession"`
		MCPCapabilities     json.RawMessage `json:"mcpCapabilities"`
		PromptCapabilities  json.RawMessage `json:"promptCapabilities"`
		SessionCapabilities json.RawMessage `json:"sessionCapabilities"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Auth) > 0 {
		unmarshalDefault(raw.Auth, &decoded.Auth)
	}
	if len(raw.LoadSession) > 0 {
		unmarshalDefault(raw.LoadSession, &decoded.LoadSession)
	}
	if len(raw.MCPCapabilities) > 0 {
		unmarshalDefault(raw.MCPCapabilities, &decoded.MCPCapabilities)
	}
	if len(raw.PromptCapabilities) > 0 {
		unmarshalDefault(raw.PromptCapabilities, &decoded.PromptCapabilities)
	}
	if len(raw.SessionCapabilities) > 0 {
		unmarshalDefault(raw.SessionCapabilities, &decoded.SessionCapabilities)
	}
	*c = AgentCapabilities(decoded)
	return nil
}

// Annotations: Optional annotations for the client. The client can use annotations to inform how objects are used or displayed
type Annotations struct {
	Meta         Meta     `json:"_meta,omitzero"`
	Audience     *[]Role  `json:"audience,omitempty"`
	LastModified *string  `json:"lastModified,omitempty"`
	Priority     *float64 `json:"priority,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *Annotations) UnmarshalJSON(data []byte) error {
	type alias Annotations
	decoded := alias{}
	raw := struct {
		Audience     json.RawMessage `json:"audience"`
		LastModified json.RawMessage `json:"lastModified"`
		Priority     json.RawMessage `json:"priority"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Audience) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.Audience, &values); err == nil && values != nil {
			items := []Role{}
			for _, value := range values {
				var item Role
				if err := json.Unmarshal(value, &item); err == nil {
					switch item {
					case RoleAssistant, RoleUser:
						items = append(items, item)
					}
				}
			}
			decoded.Audience = &items
		}
	}
	if len(raw.LastModified) > 0 {
		unmarshalDefault(raw.LastModified, &decoded.LastModified)
	}
	if len(raw.Priority) > 0 {
		unmarshalDefault(raw.Priority, &decoded.Priority)
	}
	*a = Annotations(decoded)
	return nil
}

// AudioContent: Audio provided to or from an LLM.
type AudioContent struct {
	Meta        Meta         `json:"_meta,omitzero"`
	Annotations *Annotations `json:"annotations,omitempty"`
	Data        string       `json:"data"`
	MIMEType    string       `json:"mimeType"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *AudioContent) UnmarshalJSON(data []byte) error {
	type alias AudioContent
	decoded := alias{}
	raw := struct {
		Annotations json.RawMessage `json:"annotations"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Annotations) > 0 {
		unmarshalDefault(raw.Annotations, &decoded.Annotations)
	}
	if err := requireJSONFieldsOnly(data, "data", "mimeType"); err != nil {
		return err
	}
	*c = AudioContent(decoded)
	return nil
}

// AuthMethod: Describes an available authentication method.
//
// The `type` field acts as the discriminator in the serialized JSON form.
// When no `type` is present, the method is treated as `agent`.
type AuthMethod struct {
	Meta        Meta         `json:"_meta,omitzero"`
	Description *string      `json:"description,omitempty"`
	ID          AuthMethodID `json:"id"`
	Name        string       `json:"name"`
}

// AgentAuthMethod creates an AuthMethod variant: Agent handles authentication itself through `authenticate`.
//
// This is the default when no `type` is specified.
func AgentAuthMethod(id AuthMethodID, name string) AuthMethod {
	return AuthMethod{
		ID:   id,
		Name: name,
	}
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *AuthMethod) UnmarshalJSON(data []byte) error {
	type alias AuthMethod
	decoded := alias{}
	raw := struct {
		Description json.RawMessage `json:"description"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Description) > 0 {
		unmarshalDefault(raw.Description, &decoded.Description)
	}
	if err := requireJSONFieldsOnly(data, "id", "name"); err != nil {
		return err
	}
	*m = AuthMethod(decoded)
	return nil
}

// AuthMethodAgent: Agent handles authentication itself through `authenticate`.
//
// This is the default authentication method type.
type AuthMethodAgent struct {
	Meta        Meta         `json:"_meta,omitzero"`
	Description *string      `json:"description,omitempty"`
	ID          AuthMethodID `json:"id"`
	Name        string       `json:"name"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *AuthMethodAgent) UnmarshalJSON(data []byte) error {
	type alias AuthMethodAgent
	decoded := alias{}
	raw := struct {
		Description json.RawMessage `json:"description"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Description) > 0 {
		unmarshalDefault(raw.Description, &decoded.Description)
	}
	*a = AuthMethodAgent(decoded)
	return nil
}

// AuthMethodID: Typed identifier used for auth method values on the wire.
type AuthMethodID string

// AuthenticateRequest: Request parameters for the authenticate method.
//
// Specifies which authentication method to use.
type AuthenticateRequest struct {
	Meta     Meta         `json:"_meta,omitzero"`
	MethodID AuthMethodID `json:"methodId"`
}

// AuthenticateResponse: Response to the `authenticate` method.
type AuthenticateResponse struct {
	Meta Meta `json:"_meta,omitzero"`
}

// AvailableCommand: Information about a command.
type AvailableCommand struct {
	Meta        Meta                   `json:"_meta,omitzero"`
	Description string                 `json:"description"`
	Input       *AvailableCommandInput `json:"input,omitempty"`
	Name        string                 `json:"name"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *AvailableCommand) UnmarshalJSON(data []byte) error {
	type alias AvailableCommand
	decoded := alias{}
	raw := struct {
		Input json.RawMessage `json:"input"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Input) > 0 {
		unmarshalDefault(raw.Input, &decoded.Input)
	}
	if err := requireJSONFieldsOnly(data, "name", "description"); err != nil {
		return err
	}
	*c = AvailableCommand(decoded)
	return nil
}

// AvailableCommandInput: The input specification for a command.
type AvailableCommandInput struct {
	Meta Meta   `json:"_meta,omitzero"`
	Hint string `json:"hint"`
}

// UnstructuredAvailableCommandInput creates an AvailableCommandInput variant: All text that was typed after the command name is provided as input.
func UnstructuredAvailableCommandInput(hint string) AvailableCommandInput {
	return AvailableCommandInput{
		Hint: hint,
	}
}

// AvailableCommandsUpdate: Available commands are ready or have changed
type AvailableCommandsUpdate struct {
	Meta              Meta               `json:"_meta,omitzero"`
	AvailableCommands []AvailableCommand `json:"availableCommands"`
}

// MarshalJSON implements json.Marshaler.
func (u AvailableCommandsUpdate) MarshalJSON() ([]byte, error) {
	type alias AvailableCommandsUpdate
	a := alias(u)
	if a.AvailableCommands == nil {
		a.AvailableCommands = []AvailableCommand{}
	}
	return json.Marshal(a)
}

// UnmarshalJSON implements json.Unmarshaler.
func (u *AvailableCommandsUpdate) UnmarshalJSON(data []byte) error {
	type alias AvailableCommandsUpdate
	decoded := alias{}
	raw := struct {
		AvailableCommands json.RawMessage `json:"availableCommands"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.AvailableCommands) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.AvailableCommands, &values); err == nil {
			decoded.AvailableCommands = []AvailableCommand{}
			for _, value := range values {
				var item AvailableCommand
				if err := json.Unmarshal(value, &item); err == nil {
					decoded.AvailableCommands = append(decoded.AvailableCommands, item)
				}
			}
		}
	}
	*u = AvailableCommandsUpdate(decoded)
	return nil
}

// BlobResourceContents: Binary resource contents.
type BlobResourceContents struct {
	Meta     Meta    `json:"_meta,omitzero"`
	Blob     string  `json:"blob"`
	MIMEType *string `json:"mimeType,omitempty"`
	URI      string  `json:"uri"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *BlobResourceContents) UnmarshalJSON(data []byte) error {
	type alias BlobResourceContents
	decoded := alias{}
	raw := struct {
		MIMEType json.RawMessage `json:"mimeType"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.MIMEType) > 0 {
		unmarshalDefault(raw.MIMEType, &decoded.MIMEType)
	}
	if err := requireJSONFieldsOnly(data, "blob", "uri"); err != nil {
		return err
	}
	*c = BlobResourceContents(decoded)
	return nil
}

// BooleanConfigOptionCapabilities: Capabilities for boolean session configuration options.
//
// Supplying `{}` means the client supports boolean session configuration options.
type BooleanConfigOptionCapabilities struct {
	Meta Meta `json:"_meta,omitzero"`
}

// BooleanPropertySchema: Schema for boolean properties in an elicitation form.
type BooleanPropertySchema struct {
	Meta        Meta    `json:"_meta,omitzero"`
	Default     *bool   `json:"default,omitempty"`
	Description *string `json:"description,omitempty"`
	Title       *string `json:"title,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *BooleanPropertySchema) UnmarshalJSON(data []byte) error {
	type alias BooleanPropertySchema
	decoded := alias{}
	raw := struct {
		Default     json.RawMessage `json:"default"`
		Description json.RawMessage `json:"description"`
		Title       json.RawMessage `json:"title"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Default) > 0 {
		unmarshalDefault(raw.Default, &decoded.Default)
	}
	if len(raw.Description) > 0 {
		unmarshalDefault(raw.Description, &decoded.Description)
	}
	if len(raw.Title) > 0 {
		unmarshalDefault(raw.Title, &decoded.Title)
	}
	*s = BooleanPropertySchema(decoded)
	return nil
}

// CancelNotification: Notification to cancel ongoing operations for a session.
//
// See protocol docs: [Cancellation](https://agentclientprotocol.com/protocol/prompt-turn#cancellation)
type CancelNotification struct {
	Meta      Meta      `json:"_meta,omitzero"`
	SessionID SessionID `json:"sessionId"`
}

// ClientCapabilities: Capabilities supported by the client.
//
// Advertised during initialization to inform the agent about
// available features and methods.
//
// See protocol docs: [Client Capabilities](https://agentclientprotocol.com/protocol/initialization#client-capabilities)
type ClientCapabilities struct {
	Meta        Meta                       `json:"_meta,omitzero"`
	Elicitation *ElicitationCapabilities   `json:"elicitation,omitempty"`
	Fs          *FileSystemCapabilities    `json:"fs,omitempty"`
	Session     *ClientSessionCapabilities `json:"session,omitempty"`
	Terminal    bool                       `json:"terminal,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *ClientCapabilities) UnmarshalJSON(data []byte) error {
	type alias ClientCapabilities
	decoded := alias{}
	raw := struct {
		Elicitation json.RawMessage `json:"elicitation"`
		Fs          json.RawMessage `json:"fs"`
		Session     json.RawMessage `json:"session"`
		Terminal    json.RawMessage `json:"terminal"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Elicitation) > 0 {
		unmarshalDefault(raw.Elicitation, &decoded.Elicitation)
	}
	if len(raw.Fs) > 0 {
		unmarshalDefault(raw.Fs, &decoded.Fs)
	}
	if len(raw.Session) > 0 {
		unmarshalDefault(raw.Session, &decoded.Session)
	}
	if len(raw.Terminal) > 0 {
		unmarshalDefault(raw.Terminal, &decoded.Terminal)
	}
	*c = ClientCapabilities(decoded)
	return nil
}

// ClientSessionCapabilities: Session-related capabilities supported by the client.
type ClientSessionCapabilities struct {
	Meta          Meta                              `json:"_meta,omitzero"`
	ConfigOptions *SessionConfigOptionsCapabilities `json:"configOptions,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *ClientSessionCapabilities) UnmarshalJSON(data []byte) error {
	type alias ClientSessionCapabilities
	decoded := alias{}
	raw := struct {
		ConfigOptions json.RawMessage `json:"configOptions"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.ConfigOptions) > 0 {
		unmarshalDefault(raw.ConfigOptions, &decoded.ConfigOptions)
	}
	*c = ClientSessionCapabilities(decoded)
	return nil
}

// CloseSessionRequest: Request parameters for closing an active session.
//
// If supported, the agent **must** cancel any ongoing work related to the session
// (treat it as if `session/cancel` was called) and then free up any resources
// associated with the session.
//
// Only available if the Agent supports the `sessionCapabilities.close` capability.
type CloseSessionRequest struct {
	Meta      Meta      `json:"_meta,omitzero"`
	SessionID SessionID `json:"sessionId"`
}

// CloseSessionResponse: Response from closing a session.
type CloseSessionResponse struct {
	Meta Meta `json:"_meta,omitzero"`
}

// CompleteElicitationNotification: Notification sent by the agent when a URL-based elicitation is complete.
type CompleteElicitationNotification struct {
	Meta          Meta          `json:"_meta,omitzero"`
	ElicitationID ElicitationID `json:"elicitationId"`
}

// ConfigOptionUpdate: Session configuration options have been updated.
type ConfigOptionUpdate struct {
	Meta          Meta                  `json:"_meta,omitzero"`
	ConfigOptions []SessionConfigOption `json:"configOptions"`
}

// MarshalJSON implements json.Marshaler.
func (u ConfigOptionUpdate) MarshalJSON() ([]byte, error) {
	type alias ConfigOptionUpdate
	a := alias(u)
	if a.ConfigOptions == nil {
		a.ConfigOptions = []SessionConfigOption{}
	}
	return json.Marshal(a)
}

// UnmarshalJSON implements json.Unmarshaler.
func (u *ConfigOptionUpdate) UnmarshalJSON(data []byte) error {
	type alias ConfigOptionUpdate
	decoded := alias{}
	raw := struct {
		ConfigOptions json.RawMessage `json:"configOptions"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.ConfigOptions) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.ConfigOptions, &values); err == nil {
			decoded.ConfigOptions = []SessionConfigOption{}
			for _, value := range values {
				var item SessionConfigOption
				if err := json.Unmarshal(value, &item); err == nil {
					switch item.Type {
					case SessionConfigOptionTypeSelect, SessionConfigOptionTypeBoolean:
						decoded.ConfigOptions = append(decoded.ConfigOptions, item)
					}
				}
			}
		}
	}
	*u = ConfigOptionUpdate(decoded)
	return nil
}

// Content: Standard content block (text, images, resources).
type Content struct {
	Meta    Meta         `json:"_meta,omitzero"`
	Content ContentBlock `json:"content"`
}

// ContentBlock: Content blocks represent displayable information in the Agent Client Protocol.
//
// They provide a structured way to handle various types of user-facing content—whether
// it's text from language models, images for analysis, or embedded resources for context.
//
// Content blocks appear in:
// - User prompts sent via `session/prompt`
// - Language model output streamed through `session/update` notifications
// - Progress updates and results from tool calls
//
// This structure is compatible with the Model Context Protocol (MCP), enabling
// agents to seamlessly forward content from MCP tool outputs without transformation.
//
// See protocol docs: [Content](https://agentclientprotocol.com/protocol/content)
type ContentBlock struct {
	Type        ContentBlockType         `json:"type"`
	Meta        Meta                     `json:"_meta,omitzero"`
	Annotations *Annotations             `json:"annotations,omitempty"`
	Data        string                   `json:"data,omitempty"`
	Description *string                  `json:"description,omitempty"`
	MIMEType    *string                  `json:"mimeType,omitempty"`
	Name        string                   `json:"name,omitempty"`
	Resource    EmbeddedResourceContents `json:"resource,omitzero"`
	Size        *int64                   `json:"size,omitempty"`
	Text        string                   `json:"text,omitempty"`
	Title       *string                  `json:"title,omitempty"`
	URI         *string                  `json:"uri,omitempty"`
}

// ContentBlockType is the discriminator for ContentBlock variants.
type ContentBlockType string

const (
	ContentBlockTypeText         ContentBlockType = "text"
	ContentBlockTypeImage        ContentBlockType = "image"
	ContentBlockTypeAudio        ContentBlockType = "audio"
	ContentBlockTypeResourceLink ContentBlockType = "resource_link"
	ContentBlockTypeResource     ContentBlockType = "resource"
)

// TextContentBlock creates an ContentBlock variant: Text content. May be plain text or formatted with Markdown.
//
// All agents MUST support text content blocks in prompts.
// Clients SHOULD render this text as Markdown.
func TextContentBlock(text string) ContentBlock {
	return ContentBlock{
		Type: ContentBlockTypeText,
		Text: text,
	}
}

// ImageContentBlock creates an ContentBlock variant: Images for visual context or analysis.
//
// Requires the `image` prompt capability when included in prompts.
func ImageContentBlock(data string, mimeType string) ContentBlock {
	return ContentBlock{
		Type:     ContentBlockTypeImage,
		Data:     data,
		MIMEType: &mimeType,
	}
}

// AudioContentBlock creates an ContentBlock variant: Audio data for transcription or analysis.
//
// Requires the `audio` prompt capability when included in prompts.
func AudioContentBlock(data string, mimeType string) ContentBlock {
	return ContentBlock{
		Type:     ContentBlockTypeAudio,
		Data:     data,
		MIMEType: &mimeType,
	}
}

// ResourceLinkContentBlock creates an ContentBlock variant: References to resources that the agent can access.
//
// All agents MUST support resource links in prompts.
func ResourceLinkContentBlock(name string, uri string) ContentBlock {
	return ContentBlock{
		Type: ContentBlockTypeResourceLink,
		Name: name,
		URI:  &uri,
	}
}

// ResourceContentBlock creates an ContentBlock variant: Complete resource contents embedded directly in the message.
//
// Preferred for including context as it avoids extra round-trips.
//
// Requires the `embeddedContext` prompt capability when included in prompts.
func ResourceContentBlock(resource EmbeddedResourceContents) ContentBlock {
	return ContentBlock{
		Type:     ContentBlockTypeResource,
		Resource: resource,
	}
}

// MarshalJSON implements json.Marshaler.
func (b ContentBlock) MarshalJSON() ([]byte, error) {
	type alias ContentBlock
	// The alias needs no required-field overrides when omitempty keeps every
	// required value; avoiding the pointer wrapper materially reduces copying.
	switch b.Type {
	case ContentBlockTypeText:
		if b.Text != "" {
			return json.Marshal((*alias)(&b))
		}
	case ContentBlockTypeImage, ContentBlockTypeAudio:
		if b.Data != "" && b.MIMEType != nil {
			return json.Marshal((*alias)(&b))
		}
	case ContentBlockTypeResourceLink:
		if b.Name != "" && b.URI != nil {
			return json.Marshal((*alias)(&b))
		}
	case ContentBlockTypeResource:
		if !reflect.ValueOf(b.Resource).IsZero() {
			return json.Marshal((*alias)(&b))
		}
	}
	type wire struct {
		*alias
		Text     *string                   `json:"text,omitempty"`
		Data     *string                   `json:"data,omitempty"`
		MIMEType **string                  `json:"mimeType,omitempty"`
		Name     *string                   `json:"name,omitempty"`
		URI      **string                  `json:"uri,omitempty"`
		Resource *EmbeddedResourceContents `json:"resource,omitempty"`
	}
	w := wire{alias: (*alias)(&b)}
	if !reflect.ValueOf(b.Text).IsZero() {
		Text := b.Text
		w.Text = &Text
	}
	if !reflect.ValueOf(b.Data).IsZero() {
		Data := b.Data
		w.Data = &Data
	}
	if b.MIMEType != nil {
		MIMEType := b.MIMEType
		w.MIMEType = &MIMEType
	}
	if !reflect.ValueOf(b.Name).IsZero() {
		Name := b.Name
		w.Name = &Name
	}
	if b.URI != nil {
		URI := b.URI
		w.URI = &URI
	}
	if !reflect.ValueOf(b.Resource).IsZero() {
		Resource := b.Resource
		w.Resource = &Resource
	}
	switch b.Type {
	case ContentBlockTypeText:
		Text := b.Text
		w.Text = &Text
	case ContentBlockTypeImage:
		Data := b.Data
		w.Data = &Data
		MIMEType := b.MIMEType
		w.MIMEType = &MIMEType
	case ContentBlockTypeAudio:
		Data := b.Data
		w.Data = &Data
		MIMEType := b.MIMEType
		w.MIMEType = &MIMEType
	case ContentBlockTypeResourceLink:
		Name := b.Name
		w.Name = &Name
		URI := b.URI
		w.URI = &URI
	case ContentBlockTypeResource:
		Resource := b.Resource
		w.Resource = &Resource
	}
	return json.Marshal(w)
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *ContentBlock) UnmarshalJSON(data []byte) error {
	type alias ContentBlock
	decoded := alias{}
	raw := struct {
		Annotations json.RawMessage `json:"annotations"`
		Description json.RawMessage `json:"description"`
		MIMEType    json.RawMessage `json:"mimeType"`
		Size        json.RawMessage `json:"size"`
		Title       json.RawMessage `json:"title"`
		URI         json.RawMessage `json:"uri"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Annotations) > 0 {
		unmarshalDefault(raw.Annotations, &decoded.Annotations)
	}
	if len(raw.Description) > 0 {
		unmarshalDefault(raw.Description, &decoded.Description)
	}
	if len(raw.MIMEType) > 0 {
		unmarshalDefault(raw.MIMEType, &decoded.MIMEType)
	}
	if len(raw.Size) > 0 {
		unmarshalDefault(raw.Size, &decoded.Size)
	}
	if len(raw.Title) > 0 {
		unmarshalDefault(raw.Title, &decoded.Title)
	}
	if len(raw.URI) > 0 {
		unmarshalDefault(raw.URI, &decoded.URI)
	}
	switch decoded.Type {
	case ContentBlockTypeText:
		if err := requireJSONFieldsOnly(data, "type", "text"); err != nil {
			return err
		}
	case ContentBlockTypeImage, ContentBlockTypeAudio:
		if err := requireJSONFieldsOnly(data, "type", "data", "mimeType"); err != nil {
			return err
		}
		var mimeType string
		if err := json.Unmarshal(raw.MIMEType, &mimeType); err != nil {
			return err
		}
		decoded.MIMEType = &mimeType
	case ContentBlockTypeResourceLink:
		if err := requireJSONFieldsOnly(data, "type", "name", "uri"); err != nil {
			return err
		}
		var uri string
		if err := json.Unmarshal(raw.URI, &uri); err != nil {
			return err
		}
		decoded.URI = &uri
	case ContentBlockTypeResource:
		if err := requireJSONFieldsOnly(data, "type", "resource"); err != nil {
			return err
		}
	default:
		if err := requireJSONFieldsOnly(data, "type"); err != nil {
			return err
		}
		return invalidDiscriminator("type", decoded.Type)
	}
	*b = ContentBlock(decoded)
	return nil
}

// ContentChunk: A streamed item of content
type ContentChunk struct {
	Meta      Meta         `json:"_meta,omitzero"`
	Content   ContentBlock `json:"content"`
	MessageID *MessageID   `json:"messageId,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *ContentChunk) UnmarshalJSON(data []byte) error {
	type alias ContentChunk
	decoded := alias{}
	raw := struct {
		MessageID json.RawMessage `json:"messageId"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.MessageID) > 0 {
		unmarshalDefault(raw.MessageID, &decoded.MessageID)
	}
	if err := requireJSONFieldsOnly(data, "content"); err != nil {
		return err
	}
	*c = ContentChunk(decoded)
	return nil
}

// Cost: Cost information for a session.
type Cost struct {
	Meta     Meta    `json:"_meta,omitzero"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// CreateElicitationRequest: Request from the agent to elicit structured user input.
//
// The agent sends this to the client to request information from the user,
// either via a form or by directing them to a URL.
// Elicitations are tied to a session (optionally a tool call) or a request.
type CreateElicitationRequest struct {
	Mode            CreateElicitationRequestType `json:"mode"`
	Meta            Meta                         `json:"_meta,omitzero"`
	Fields          map[string]json.RawMessage   `json:"-"`
	ElicitationID   ElicitationID                `json:"elicitationId,omitempty"`
	Message         string                       `json:"message"`
	RequestID       RequestID                    `json:"requestId,omitempty"`
	RequestedSchema ElicitationSchema            `json:"requestedSchema,omitzero"`
	SessionID       SessionID                    `json:"sessionId,omitempty"`
	ToolCallID      *ToolCallID                  `json:"toolCallId,omitempty"`
	URL             string                       `json:"url,omitempty"`
}

// CreateElicitationRequestType is the discriminator for CreateElicitationRequest variants.
type CreateElicitationRequestType string

const (
	CreateElicitationRequestTypeForm CreateElicitationRequestType = "form"
	CreateElicitationRequestTypeURL  CreateElicitationRequestType = "url"
)

// SessionFormCreateElicitationRequest creates a session-scoped form elicitation.
func SessionFormCreateElicitationRequest(message string, requestedSchema ElicitationSchema, sessionID SessionID) CreateElicitationRequest {
	return CreateElicitationRequest{
		Mode:            CreateElicitationRequestTypeForm,
		Message:         message,
		RequestedSchema: requestedSchema,
		SessionID:       sessionID,
	}
}

// RequestFormCreateElicitationRequest creates a request-scoped form elicitation.
func RequestFormCreateElicitationRequest(message string, requestedSchema ElicitationSchema, requestID RequestID) CreateElicitationRequest {
	if len(requestID) == 0 {
		requestID = json.RawMessage("null")
	}
	return CreateElicitationRequest{
		Mode:            CreateElicitationRequestTypeForm,
		Message:         message,
		RequestID:       requestID,
		RequestedSchema: requestedSchema,
	}
}

// SessionURLCreateElicitationRequest creates a session-scoped URL elicitation.
func SessionURLCreateElicitationRequest(message string, elicitationID ElicitationID, url string, sessionID SessionID) CreateElicitationRequest {
	return CreateElicitationRequest{
		Mode:          CreateElicitationRequestTypeURL,
		ElicitationID: elicitationID,
		Message:       message,
		SessionID:     sessionID,
		URL:           url,
	}
}

// RequestURLCreateElicitationRequest creates a request-scoped URL elicitation.
func RequestURLCreateElicitationRequest(message string, elicitationID ElicitationID, url string, requestID RequestID) CreateElicitationRequest {
	if len(requestID) == 0 {
		requestID = json.RawMessage("null")
	}
	return CreateElicitationRequest{
		Mode:          CreateElicitationRequestTypeURL,
		ElicitationID: elicitationID,
		Message:       message,
		RequestID:     requestID,
		URL:           url,
	}
}

// SessionOtherCreateElicitationRequest creates a session-scoped custom or future elicitation mode.
//
// Values beginning with `_` are reserved for implementation-specific
// extensions. Unknown values that do not begin with `_` are reserved for
// future ACP variants.
//
// Clients that do not understand this mode should preserve the raw payload
// when storing, replaying, proxying, or forwarding elicitation requests.
// They MUST NOT render it as a known elicitation mode.
func SessionOtherCreateElicitationRequest(message string, mode string, sessionID SessionID, fields map[string]json.RawMessage) CreateElicitationRequest {
	return CreateElicitationRequest{
		Fields:    fields,
		Message:   message,
		Mode:      CreateElicitationRequestType(mode),
		SessionID: sessionID,
	}
}

// RequestOtherCreateElicitationRequest creates a request-scoped custom or future elicitation mode.
func RequestOtherCreateElicitationRequest(message string, mode string, requestID RequestID, fields map[string]json.RawMessage) CreateElicitationRequest {
	if len(requestID) == 0 {
		requestID = json.RawMessage("null")
	}
	return CreateElicitationRequest{
		Fields:    fields,
		Message:   message,
		Mode:      CreateElicitationRequestType(mode),
		RequestID: requestID,
	}
}

// MarshalJSON implements json.Marshaler.
func (r CreateElicitationRequest) MarshalJSON() ([]byte, error) {
	type alias CreateElicitationRequest
	type wire struct {
		*alias
		RequestedSchema *ElicitationSchema `json:"requestedSchema,omitempty"`
		ElicitationID   *ElicitationID     `json:"elicitationId,omitempty"`
		URL             *string            `json:"url,omitempty"`
	}
	value := alias(r)
	if r.SessionID != "" {
		value.RequestID = nil
	}
	w := wire{alias: &value}
	if !reflect.ValueOf(r.RequestedSchema).IsZero() {
		RequestedSchema := r.RequestedSchema
		w.RequestedSchema = &RequestedSchema
	}
	if !reflect.ValueOf(r.ElicitationID).IsZero() {
		elicitationID := r.ElicitationID
		w.ElicitationID = &elicitationID
	}
	if !reflect.ValueOf(r.URL).IsZero() {
		URL := r.URL
		w.URL = &URL
	}
	switch r.Mode {
	case CreateElicitationRequestTypeForm:
		RequestedSchema := r.RequestedSchema
		w.RequestedSchema = &RequestedSchema
	case CreateElicitationRequestTypeURL:
		elicitationID := r.ElicitationID
		w.ElicitationID = &elicitationID
		URL := r.URL
		w.URL = &URL
	}
	fields := r.Fields
	if r.SessionID == "" && len(r.RequestID) == 0 {
		fields = make(map[string]json.RawMessage, len(r.Fields)+1)
		maps.Copy(fields, r.Fields)
		fields["sessionId"] = json.RawMessage(`""`)
	}
	return marshalJSONWithFields(w, fields)
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *CreateElicitationRequest) UnmarshalJSON(data []byte) error {
	fields, err := requireJSONFields(data, "mode", "message")
	if err != nil {
		return err
	}
	sessionScoped, err := elicitationScopeIsSession(fields)
	if err != nil {
		return err
	}

	var mode CreateElicitationRequestType
	if err := json.Unmarshal(fields["mode"], &mode); err != nil {
		return err
	}
	if mode != CreateElicitationRequestTypeForm && mode != CreateElicitationRequestTypeURL {
		var decoded CreateElicitationRequest
		decoded.Mode = mode
		if err := json.Unmarshal(fields["message"], &decoded.Message); err != nil {
			return err
		}
		if raw := fields["_meta"]; len(raw) > 0 {
			unmarshalDefault(raw, &decoded.Meta)
		}
		if sessionScoped {
			raw := fields["sessionId"]
			if err := json.Unmarshal(raw, &decoded.SessionID); err != nil {
				return err
			}
			if raw := fields["toolCallId"]; len(raw) > 0 {
				unmarshalDefault(raw, &decoded.ToolCallID)
			}
		} else {
			raw := fields["requestId"]
			decoded.RequestID = append(decoded.RequestID[:0], raw...)
		}
		delete(fields, "mode")
		delete(fields, "message")
		delete(fields, "_meta")
		delete(fields, "sessionId")
		delete(fields, "requestId")
		delete(fields, "toolCallId")
		if len(fields) > 0 {
			decoded.Fields = fields
		}
		*r = decoded
		return nil
	}

	type alias CreateElicitationRequest
	decoded := alias{}
	raw := struct {
		ToolCallID json.RawMessage `json:"toolCallId"`
		RequestID  json.RawMessage `json:"requestId"`
		SessionID  json.RawMessage `json:"sessionId"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if sessionScoped {
		if err := json.Unmarshal(raw.SessionID, &decoded.SessionID); err != nil {
			return err
		}
		if len(raw.ToolCallID) > 0 {
			unmarshalDefault(raw.ToolCallID, &decoded.ToolCallID)
		}
	} else {
		decoded.RequestID = append(decoded.RequestID[:0], raw.RequestID...)
	}
	switch mode {
	case CreateElicitationRequestTypeForm:
		if err := requireFields(fields, "requestedSchema"); err != nil {
			return err
		}
	case CreateElicitationRequestTypeURL:
		if err := requireFields(fields, "elicitationId", "url"); err != nil {
			return err
		}
	}
	*r = CreateElicitationRequest(decoded)
	return nil
}

// CreateElicitationResponse: Response from the client to an elicitation request.
type CreateElicitationResponse struct {
	Action  CreateElicitationResponseType       `json:"action"`
	Meta    Meta                                `json:"_meta,omitzero"`
	Fields  map[string]json.RawMessage          `json:"-"`
	Content *map[string]ElicitationContentValue `json:"content,omitempty"`
}

// CreateElicitationResponseType is the discriminator for CreateElicitationResponse variants.
type CreateElicitationResponseType string

const (
	CreateElicitationResponseTypeAccept  CreateElicitationResponseType = "accept"
	CreateElicitationResponseTypeDecline CreateElicitationResponseType = "decline"
	CreateElicitationResponseTypeCancel  CreateElicitationResponseType = "cancel"
)

// AcceptCreateElicitationResponse creates a CreateElicitationResponse accept variant.
func AcceptCreateElicitationResponse() CreateElicitationResponse {
	return CreateElicitationResponse{
		Action: CreateElicitationResponseTypeAccept,
	}
}

// DeclineCreateElicitationResponse creates a CreateElicitationResponse decline variant.
func DeclineCreateElicitationResponse() CreateElicitationResponse {
	return CreateElicitationResponse{
		Action: CreateElicitationResponseTypeDecline,
	}
}

// CancelCreateElicitationResponse creates a CreateElicitationResponse cancel variant.
func CancelCreateElicitationResponse() CreateElicitationResponse {
	return CreateElicitationResponse{
		Action: CreateElicitationResponseTypeCancel,
	}
}

// OtherCreateElicitationResponse creates a custom or future CreateElicitationResponse action.
//
// Values beginning with `_` are reserved for implementation-specific
// extensions. Unknown values that do not begin with `_` are reserved for
// future ACP variants.
//
// Agents that do not understand this action should preserve the raw
// payload when storing, replaying, proxying, or forwarding elicitation
// responses. They MUST NOT treat it as a known elicitation action.
func OtherCreateElicitationResponse(action string, fields map[string]json.RawMessage) CreateElicitationResponse {
	return CreateElicitationResponse{
		Action: CreateElicitationResponseType(action),
		Fields: fields,
	}
}

// MarshalJSON implements json.Marshaler.
func (r CreateElicitationResponse) MarshalJSON() ([]byte, error) {
	type alias CreateElicitationResponse
	return marshalJSONWithFields(alias(r), r.Fields)
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *CreateElicitationResponse) UnmarshalJSON(data []byte) error {
	fields, err := requireJSONFields(data, "action")
	if err != nil {
		return err
	}
	var action CreateElicitationResponseType
	if err := json.Unmarshal(fields["action"], &action); err != nil {
		return err
	}

	if action == CreateElicitationResponseTypeAccept {
		type alias CreateElicitationResponse
		var decoded alias
		if err := json.Unmarshal(data, &decoded); err != nil {
			return err
		}
		*r = CreateElicitationResponse(decoded)
		return nil
	}

	decoded := CreateElicitationResponse{Action: action}
	if raw := fields["_meta"]; len(raw) > 0 {
		unmarshalDefault(raw, &decoded.Meta)
	}
	if action != CreateElicitationResponseTypeDecline && action != CreateElicitationResponseTypeCancel {
		delete(fields, "action")
		delete(fields, "_meta")
		if len(fields) > 0 {
			decoded.Fields = fields
		}
	}
	*r = decoded
	return nil
}

// CreateTerminalRequest: Request to create a new terminal and execute a command.
type CreateTerminalRequest struct {
	Meta            Meta          `json:"_meta,omitzero"`
	Args            []string      `json:"args,omitempty"`
	Command         string        `json:"command"`
	Cwd             *string       `json:"cwd,omitempty"`
	Env             []EnvVariable `json:"env,omitempty"`
	OutputByteLimit *uint64       `json:"outputByteLimit,omitempty"`
	SessionID       SessionID     `json:"sessionId"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *CreateTerminalRequest) UnmarshalJSON(data []byte) error {
	type alias CreateTerminalRequest
	decoded := alias{}
	raw := struct {
		Args            json.RawMessage `json:"args"`
		Cwd             json.RawMessage `json:"cwd"`
		Env             json.RawMessage `json:"env"`
		OutputByteLimit json.RawMessage `json:"outputByteLimit"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Args) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.Args, &values); err == nil {
			decoded.Args = []string{}
			for _, value := range values {
				var item string
				if err := json.Unmarshal(value, &item); err == nil {
					decoded.Args = append(decoded.Args, item)
				}
			}
		}
	}
	if len(raw.Cwd) > 0 {
		unmarshalDefault(raw.Cwd, &decoded.Cwd)
	}
	if len(raw.Env) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.Env, &values); err == nil {
			decoded.Env = []EnvVariable{}
			for _, value := range values {
				var item EnvVariable
				if err := json.Unmarshal(value, &item); err == nil {
					decoded.Env = append(decoded.Env, item)
				}
			}
		}
	}
	if len(raw.OutputByteLimit) > 0 {
		unmarshalDefault(raw.OutputByteLimit, &decoded.OutputByteLimit)
	}
	*r = CreateTerminalRequest(decoded)
	return nil
}

// CreateTerminalResponse: Response containing the ID of the created terminal.
type CreateTerminalResponse struct {
	Meta       Meta       `json:"_meta,omitzero"`
	TerminalID TerminalID `json:"terminalId"`
}

// CurrentModeUpdate: The current mode of the session has changed
//
// See protocol docs: [Session Modes](https://agentclientprotocol.com/protocol/session-modes)
type CurrentModeUpdate struct {
	Meta          Meta          `json:"_meta,omitzero"`
	CurrentModeID SessionModeID `json:"currentModeId"`
}

// DeleteSessionRequest: Request parameters for deleting an existing session from `session/list`.
//
// Only available if the Agent supports the `sessionCapabilities.delete` capability.
type DeleteSessionRequest struct {
	Meta      Meta      `json:"_meta,omitzero"`
	SessionID SessionID `json:"sessionId"`
}

// DeleteSessionResponse: Response from deleting a session.
type DeleteSessionResponse struct {
	Meta Meta `json:"_meta,omitzero"`
}

// Diff: A diff representing file modifications.
//
// Shows changes to files in a format suitable for display in the client UI.
//
// See protocol docs: [Content](https://agentclientprotocol.com/protocol/tool-calls#content)
type Diff struct {
	Meta    Meta    `json:"_meta,omitzero"`
	NewText string  `json:"newText"`
	OldText *string `json:"oldText,omitempty"`
	Path    string  `json:"path"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *Diff) UnmarshalJSON(data []byte) error {
	type alias Diff
	decoded := alias{}
	raw := struct {
		OldText json.RawMessage `json:"oldText"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.OldText) > 0 {
		unmarshalDefault(raw.OldText, &decoded.OldText)
	}
	if err := requireJSONFieldsOnly(data, "path", "newText"); err != nil {
		return err
	}
	*d = Diff(decoded)
	return nil
}

// ElicitationAcceptAction: The user accepted the elicitation and provided content.
type ElicitationAcceptAction struct {
	Content *map[string]ElicitationContentValue `json:"content,omitempty"`
}

// ElicitationCapabilities: Elicitation capabilities supported by the client.
type ElicitationCapabilities struct {
	Meta Meta                         `json:"_meta,omitzero"`
	Form *ElicitationFormCapabilities `json:"form,omitempty"`
	URL  *ElicitationURLCapabilities  `json:"url,omitempty"`
}

// SupportsForm reports whether form-based elicitation is advertised.
func (c *ElicitationCapabilities) SupportsForm() bool {
	return c != nil && c.Form != nil
}

// SupportsURL reports whether URL-based elicitation is advertised.
func (c *ElicitationCapabilities) SupportsURL() bool {
	return c != nil && c.URL != nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *ElicitationCapabilities) UnmarshalJSON(data []byte) error {
	type alias ElicitationCapabilities
	decoded := alias{}
	raw := struct {
		Form json.RawMessage `json:"form"`
		URL  json.RawMessage `json:"url"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Form) > 0 {
		unmarshalDefault(raw.Form, &decoded.Form)
	}
	if len(raw.URL) > 0 {
		unmarshalDefault(raw.URL, &decoded.URL)
	}
	*c = ElicitationCapabilities(decoded)
	return nil
}

// ElicitationContentValue: Allowed wire representations for [`ElicitationContentValue`].
type ElicitationContentValue json.RawMessage

// ElicitationFormCapabilities: Form-based elicitation capabilities.
//
// Supplying `{}` means the client supports form-based elicitation.
type ElicitationFormCapabilities struct {
	Meta Meta `json:"_meta,omitzero"`
}

// ElicitationFormMode: Form-based elicitation mode where the client renders a form from the provided schema.
type ElicitationFormMode struct {
	RequestID       *RequestID        `json:"requestId,omitempty"`
	RequestedSchema ElicitationSchema `json:"requestedSchema"`
	SessionID       *SessionID        `json:"sessionId,omitempty"`
	ToolCallID      *ToolCallID       `json:"toolCallId,omitempty"`
}

// SessionElicitationFormMode creates an ElicitationFormMode tied to a session, optionally to a specific tool call within that session.
func SessionElicitationFormMode(requestedSchema ElicitationSchema, sessionID SessionID) ElicitationFormMode {
	return ElicitationFormMode{
		RequestedSchema: requestedSchema,
		SessionID:       &sessionID,
	}
}

// RequestElicitationFormMode creates an ElicitationFormMode tied to a specific JSON-RPC request outside of a session
// (e.g., during auth/configuration phases before any session is started).
func RequestElicitationFormMode(requestedSchema ElicitationSchema, requestID RequestID) ElicitationFormMode {
	return ElicitationFormMode{
		RequestID:       &requestID,
		RequestedSchema: requestedSchema,
	}
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *ElicitationFormMode) UnmarshalJSON(data []byte) error {
	type alias ElicitationFormMode
	decoded := alias{}
	raw := struct {
		ToolCallID json.RawMessage `json:"toolCallId"`
		RequestID  json.RawMessage `json:"requestId"`
		SessionID  json.RawMessage `json:"sessionId"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	fields, err := requireJSONFields(data, "requestedSchema")
	if err != nil {
		return err
	}
	sessionScoped, err := elicitationScopeIsSession(fields)
	if err != nil {
		return err
	}
	if sessionScoped {
		var sessionID SessionID
		if err := json.Unmarshal(raw.SessionID, &sessionID); err != nil {
			return err
		}
		decoded.SessionID = &sessionID
		if len(raw.ToolCallID) > 0 {
			unmarshalDefault(raw.ToolCallID, &decoded.ToolCallID)
		}
	} else {
		requestID := append(json.RawMessage(nil), raw.RequestID...)
		decoded.RequestID = &requestID
	}
	*m = ElicitationFormMode(decoded)
	return nil
}

// ElicitationID: Unique identifier for an elicitation.
type ElicitationID string

// ElicitationPropertySchema: Property schema for elicitation form fields.
//
// Each variant corresponds to a JSON Schema `"type"` value.
// Single-select enums use the `String` variant with `enum` or `oneOf` set.
// Multi-select enums use the `Array` variant.
type ElicitationPropertySchema struct {
	Type        ElicitationPropertySchemaType `json:"type"`
	Meta        Meta                          `json:"_meta,omitzero"`
	Fields      map[string]json.RawMessage    `json:"-"`
	Default     any                           `json:"default,omitempty"`
	Description *string                       `json:"description,omitempty"`
	Enum        *[]string                     `json:"enum,omitempty"`
	Format      *StringFormat                 `json:"format,omitempty"`
	Items       MultiSelectItems              `json:"items,omitzero"`
	MaxItems    *uint64                       `json:"maxItems,omitempty"`
	MaxLength   *uint32                       `json:"maxLength,omitempty"`
	Maximum     any                           `json:"maximum,omitempty"`
	MinItems    *uint64                       `json:"minItems,omitempty"`
	MinLength   *uint32                       `json:"minLength,omitempty"`
	Minimum     any                           `json:"minimum,omitempty"`
	OneOf       *[]EnumOption                 `json:"oneOf,omitempty"`
	Pattern     *string                       `json:"pattern,omitempty"`
	Title       *string                       `json:"title,omitempty"`
}

// ElicitationPropertySchemaType is the discriminator for ElicitationPropertySchema variants.
type ElicitationPropertySchemaType string

const (
	ElicitationPropertySchemaTypeString  ElicitationPropertySchemaType = "string"
	ElicitationPropertySchemaTypeNumber  ElicitationPropertySchemaType = "number"
	ElicitationPropertySchemaTypeInteger ElicitationPropertySchemaType = "integer"
	ElicitationPropertySchemaTypeBoolean ElicitationPropertySchemaType = "boolean"
	ElicitationPropertySchemaTypeArray   ElicitationPropertySchemaType = "array"
)

// StringElicitationPropertySchema creates a string ElicitationPropertySchema.
func StringElicitationPropertySchema() ElicitationPropertySchema {
	return ElicitationPropertySchema{
		Type: ElicitationPropertySchemaTypeString,
	}
}

// NumberElicitationPropertySchema creates a number ElicitationPropertySchema.
func NumberElicitationPropertySchema() ElicitationPropertySchema {
	return ElicitationPropertySchema{
		Type: ElicitationPropertySchemaTypeNumber,
	}
}

// IntegerElicitationPropertySchema creates an integer ElicitationPropertySchema.
func IntegerElicitationPropertySchema() ElicitationPropertySchema {
	return ElicitationPropertySchema{
		Type: ElicitationPropertySchemaTypeInteger,
	}
}

// BooleanElicitationPropertySchema creates a boolean ElicitationPropertySchema.
func BooleanElicitationPropertySchema() ElicitationPropertySchema {
	return ElicitationPropertySchema{
		Type: ElicitationPropertySchemaTypeBoolean,
	}
}

// ArrayElicitationPropertySchema creates a multi-select array ElicitationPropertySchema.
func ArrayElicitationPropertySchema(items MultiSelectItems) ElicitationPropertySchema {
	return ElicitationPropertySchema{
		Type:  ElicitationPropertySchemaTypeArray,
		Items: items,
	}
}

// OtherElicitationPropertySchema creates a custom or future ElicitationPropertySchema.
//
// Values beginning with `_` are reserved for implementation-specific
// extensions. Unknown values that do not begin with `_` are reserved for
// future ACP variants.
//
// Clients that do not understand this property schema type should preserve
// the raw schema when storing, replaying, proxying, or forwarding
// elicitation requests. They MUST NOT render it as a known input control.
func OtherElicitationPropertySchema(typeName string, fields map[string]json.RawMessage) ElicitationPropertySchema {
	return ElicitationPropertySchema{
		Type:   ElicitationPropertySchemaType(typeName),
		Fields: fields,
	}
}

// MarshalJSON implements json.Marshaler.
func (s ElicitationPropertySchema) MarshalJSON() ([]byte, error) {
	type alias ElicitationPropertySchema
	type wire struct {
		*alias
		Items *MultiSelectItems `json:"items,omitempty"`
	}
	w := wire{alias: (*alias)(&s)}
	if !reflect.ValueOf(s.Items).IsZero() {
		Items := s.Items
		w.Items = &Items
	}
	switch s.Type {
	case ElicitationPropertySchemaTypeArray:
		Items := s.Items
		w.Items = &Items
	case ElicitationPropertySchemaTypeString, ElicitationPropertySchemaTypeNumber, ElicitationPropertySchemaTypeInteger, ElicitationPropertySchemaTypeBoolean:
	default:
	}
	return marshalJSONWithFields(w, s.Fields)
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *ElicitationPropertySchema) UnmarshalJSON(data []byte) error {
	fields, err := requireJSONFields(data, "type")
	if err != nil {
		return err
	}
	var schemaType ElicitationPropertySchemaType
	if err := json.Unmarshal(fields["type"], &schemaType); err != nil {
		return err
	}

	decoded := ElicitationPropertySchema{Type: schemaType}
	switch schemaType {
	case ElicitationPropertySchemaTypeString:
		var value StringPropertySchema
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		decoded.Meta, decoded.Description, decoded.Enum, decoded.Format = value.Meta, value.Description, value.Enum, value.Format
		decoded.MaxLength, decoded.MinLength, decoded.OneOf, decoded.Pattern, decoded.Title = value.MaxLength, value.MinLength, value.OneOf, value.Pattern, value.Title
		if value.Default != nil {
			decoded.Default = *value.Default
		}
	case ElicitationPropertySchemaTypeNumber:
		var value NumberPropertySchema
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		decoded.Meta, decoded.Description, decoded.Title = value.Meta, value.Description, value.Title
		if value.Default != nil {
			decoded.Default = *value.Default
		}
		if value.Maximum != nil {
			decoded.Maximum = *value.Maximum
		}
		if value.Minimum != nil {
			decoded.Minimum = *value.Minimum
		}
	case ElicitationPropertySchemaTypeInteger:
		var value IntegerPropertySchema
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		decoded.Meta, decoded.Description, decoded.Title = value.Meta, value.Description, value.Title
		if value.Default != nil {
			decoded.Default = *value.Default
		}
		if value.Maximum != nil {
			decoded.Maximum = *value.Maximum
		}
		if value.Minimum != nil {
			decoded.Minimum = *value.Minimum
		}
	case ElicitationPropertySchemaTypeBoolean:
		var value BooleanPropertySchema
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		decoded.Meta, decoded.Description, decoded.Title = value.Meta, value.Description, value.Title
		if value.Default != nil {
			decoded.Default = *value.Default
		}
	case ElicitationPropertySchemaTypeArray:
		var value MultiSelectPropertySchema
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		decoded.Meta, decoded.Description, decoded.Items = value.Meta, value.Description, value.Items
		decoded.MaxItems, decoded.MinItems, decoded.Title = value.MaxItems, value.MinItems, value.Title
		if value.Default != nil {
			decoded.Default = *value.Default
		}
	default:
		delete(fields, "type")
		decoded.Fields = fields
	}
	*s = decoded
	return nil
}

// ElicitationRequestScope: Request-scoped elicitation, tied to a specific JSON-RPC request outside of a session
// (e.g., during auth/configuration phases before any session is started).
type ElicitationRequestScope struct {
	RequestID RequestID `json:"requestId"`
}

// ElicitationSchema: Type-safe elicitation schema for requesting structured user input.
//
// This represents a JSON Schema object with primitive-typed properties,
// as required by the elicitation specification.
type ElicitationSchema struct {
	Meta        Meta                                 `json:"_meta,omitzero"`
	Description *string                              `json:"description,omitempty"`
	Properties  map[string]ElicitationPropertySchema `json:"properties,omitempty"`
	Required    *[]string                            `json:"required,omitempty"`
	Title       *string                              `json:"title,omitempty"`
	Type        ElicitationSchemaType                `json:"type,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (s ElicitationSchema) MarshalJSON() ([]byte, error) {
	type alias ElicitationSchema
	value := alias(s)
	if value.Type == "" {
		value.Type = ElicitationSchemaTypeObject
	}
	if value.Properties == nil {
		value.Properties = map[string]ElicitationPropertySchema{}
	}
	type wire struct {
		*alias
		Properties map[string]ElicitationPropertySchema `json:"properties"`
		Type       ElicitationSchemaType                `json:"type"`
	}
	return json.Marshal(wire{alias: &value, Properties: value.Properties, Type: value.Type})
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *ElicitationSchema) UnmarshalJSON(data []byte) error {
	type alias ElicitationSchema
	decoded := alias{Properties: map[string]ElicitationPropertySchema{}, Type: ElicitationSchemaTypeObject}
	raw := struct {
		Description json.RawMessage `json:"description"`
		Properties  json.RawMessage `json:"properties"`
		Title       json.RawMessage `json:"title"`
		Type        json.RawMessage `json:"type"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Description) > 0 {
		unmarshalDefault(raw.Description, &decoded.Description)
	}
	if len(raw.Properties) > 0 {
		if jsonValueIsNull(raw.Properties) {
			return errors.New("acp: elicitation schema properties must be an object")
		}
		if err := json.Unmarshal(raw.Properties, &decoded.Properties); err != nil {
			return err
		}
	}
	if len(raw.Title) > 0 {
		unmarshalDefault(raw.Title, &decoded.Title)
	}
	if len(raw.Type) > 0 {
		var value ElicitationSchemaType
		if err := json.Unmarshal(raw.Type, &value); err == nil && value == ElicitationSchemaTypeObject {
			decoded.Type = value
		}
	}
	*s = ElicitationSchema(decoded)
	return nil
}

// ElicitationSchemaType: Type discriminator for elicitation schemas.
type ElicitationSchemaType string

const (
	// ElicitationSchemaTypeObject: Object schema type.
	ElicitationSchemaTypeObject ElicitationSchemaType = "object"
)

// ElicitationSessionScope: Session-scoped elicitation, optionally tied to a specific tool call.
//
// When `tool_call_id` is set, the elicitation is tied to a specific tool call.
// This is useful when an agent receives an elicitation from an MCP server
// during a tool call and needs to redirect it to the user.
type ElicitationSessionScope struct {
	SessionID  SessionID   `json:"sessionId"`
	ToolCallID *ToolCallID `json:"toolCallId,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *ElicitationSessionScope) UnmarshalJSON(data []byte) error {
	type alias ElicitationSessionScope
	decoded := alias{}
	raw := struct {
		ToolCallID json.RawMessage `json:"toolCallId"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.ToolCallID) > 0 {
		unmarshalDefault(raw.ToolCallID, &decoded.ToolCallID)
	}
	if err := requireJSONFieldsOnly(data, "sessionId"); err != nil {
		return err
	}
	*s = ElicitationSessionScope(decoded)
	return nil
}

// ElicitationURLCapabilities: URL-based elicitation capabilities.
//
// Supplying `{}` means the client supports URL-based elicitation.
type ElicitationURLCapabilities struct {
	Meta Meta `json:"_meta,omitzero"`
}

// ElicitationURLMode: URL-based elicitation mode where the client directs the user to a URL.
type ElicitationURLMode struct {
	ElicitationID ElicitationID `json:"elicitationId"`
	RequestID     *RequestID    `json:"requestId,omitempty"`
	SessionID     *SessionID    `json:"sessionId,omitempty"`
	ToolCallID    *ToolCallID   `json:"toolCallId,omitempty"`
	URL           string        `json:"url"`
}

// SessionElicitationURLMode creates an ElicitationURLMode variant tied to a session, optionally to a specific tool call within that session.
func SessionElicitationURLMode(elicitationID ElicitationID, url string, sessionID SessionID) ElicitationURLMode {
	return ElicitationURLMode{
		ElicitationID: elicitationID,
		SessionID:     &sessionID,
		URL:           url,
	}
}

// RequestElicitationURLMode creates an ElicitationURLMode variant tied to a specific JSON-RPC request outside of a session
// (e.g., during auth/configuration phases before any session is started).
func RequestElicitationURLMode(elicitationID ElicitationID, url string, requestID RequestID) ElicitationURLMode {
	return ElicitationURLMode{
		ElicitationID: elicitationID,
		RequestID:     &requestID,
		URL:           url,
	}
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *ElicitationURLMode) UnmarshalJSON(data []byte) error {
	type alias ElicitationURLMode
	decoded := alias{}
	raw := struct {
		ToolCallID json.RawMessage `json:"toolCallId"`
		RequestID  json.RawMessage `json:"requestId"`
		SessionID  json.RawMessage `json:"sessionId"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	fields, err := requireJSONFields(data, "elicitationId", "url")
	if err != nil {
		return err
	}
	sessionScoped, err := elicitationScopeIsSession(fields)
	if err != nil {
		return err
	}
	if sessionScoped {
		var sessionID SessionID
		if err := json.Unmarshal(raw.SessionID, &sessionID); err != nil {
			return err
		}
		decoded.SessionID = &sessionID
		if len(raw.ToolCallID) > 0 {
			unmarshalDefault(raw.ToolCallID, &decoded.ToolCallID)
		}
	} else {
		requestID := append(json.RawMessage(nil), raw.RequestID...)
		decoded.RequestID = &requestID
	}
	*m = ElicitationURLMode(decoded)
	return nil
}

// EmbeddedResource: The contents of a resource, embedded into a prompt or tool call result.
type EmbeddedResource struct {
	Meta        Meta                     `json:"_meta,omitzero"`
	Annotations *Annotations             `json:"annotations,omitempty"`
	Resource    EmbeddedResourceContents `json:"resource"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *EmbeddedResource) UnmarshalJSON(data []byte) error {
	type alias EmbeddedResource
	decoded := alias{}
	raw := struct {
		Annotations json.RawMessage `json:"annotations"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Annotations) > 0 {
		unmarshalDefault(raw.Annotations, &decoded.Annotations)
	}
	if err := requireJSONFieldsOnly(data, "resource"); err != nil {
		return err
	}
	*r = EmbeddedResource(decoded)
	return nil
}

// EmbeddedResourceContents: Resource content that can be embedded in a message.
type EmbeddedResourceContents struct {
	Meta     Meta    `json:"_meta,omitzero"`
	Blob     *string `json:"blob,omitempty"`
	MIMEType *string `json:"mimeType,omitempty"`
	Text     *string `json:"text,omitempty"`
	URI      string  `json:"uri"`
}

// TextEmbeddedResourceContents creates an EmbeddedResourceContents variant: Text resource contents embedded directly in the message.
func TextEmbeddedResourceContents(text string, uri string) EmbeddedResourceContents {
	return EmbeddedResourceContents{
		Text: &text,
		URI:  uri,
	}
}

// BlobEmbeddedResourceContents creates an EmbeddedResourceContents variant: Binary resource contents embedded directly in the message.
func BlobEmbeddedResourceContents(blob string, uri string) EmbeddedResourceContents {
	return EmbeddedResourceContents{
		Blob: &blob,
		URI:  uri,
	}
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *EmbeddedResourceContents) UnmarshalJSON(data []byte) error {
	type alias EmbeddedResourceContents
	decoded := alias{}
	raw := struct {
		MIMEType json.RawMessage `json:"mimeType"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.MIMEType) > 0 {
		unmarshalDefault(raw.MIMEType, &decoded.MIMEType)
	}
	fields, err := requireJSONFields(data, "uri")
	if err != nil {
		return err
	}
	_, hasText := fields["text"]
	_, hasBlob := fields["blob"]
	if hasText == hasBlob {
		return invalidDiscriminator("resource", "text/blob")
	}
	if hasText {
		if err := requireFields(fields, "text"); err != nil {
			return err
		}
	} else if err := requireFields(fields, "blob"); err != nil {
		return err
	}
	*r = EmbeddedResourceContents(decoded)
	return nil
}

// EnumOption: A titled enum option with a const value, human-readable title, and optional description.
type EnumOption struct {
	Meta        Meta    `json:"_meta,omitzero"`
	Const       string  `json:"const"`
	Description *string `json:"description,omitempty"`
	Title       string  `json:"title"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *EnumOption) UnmarshalJSON(data []byte) error {
	type alias EnumOption
	decoded := alias{}
	raw := struct {
		Description json.RawMessage `json:"description"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Description) > 0 {
		unmarshalDefault(raw.Description, &decoded.Description)
	}
	if err := requireJSONFieldsOnly(data, "const", "title"); err != nil {
		return err
	}
	*o = EnumOption(decoded)
	return nil
}

// EnvVariable: An environment variable to set when launching an MCP server.
type EnvVariable struct {
	Meta  Meta   `json:"_meta,omitzero"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Error: JSON-RPC error object.
//
// Represents an error that occurred during method execution, following the
// JSON-RPC 2.0 error object specification with optional additional data.
//
// See protocol docs: [JSON-RPC Error Object](https://www.jsonrpc.org/specification#error_object)
type Error struct {
	Code    ErrorCode `json:"code"`
	Data    any       `json:"data,omitempty"`
	Message string    `json:"message"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *Error) UnmarshalJSON(data []byte) error {
	type alias Error
	decoded := alias{}
	raw := struct {
		Data json.RawMessage `json:"data"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Data) > 0 {
		unmarshalDefault(raw.Data, &decoded.Data)
	}
	*e = Error(decoded)
	return nil
}

// ErrorCode: Predefined error codes for common JSON-RPC and ACP-specific errors.
//
// These codes follow the JSON-RPC 2.0 specification for standard errors
// and use the reserved range (-32000 to -32099) for protocol-specific errors.
type ErrorCode int32

const (
	// ErrorCodeParseError: **Parse error**: Invalid JSON was received by the server.
	// An error occurred on the server while parsing the JSON text.
	ErrorCodeParseError ErrorCode = -32700
	// ErrorCodeInvalidRequest: **Invalid request**: The JSON sent is not a valid Request object.
	ErrorCodeInvalidRequest ErrorCode = -32600
	// ErrorCodeMethodNotFound: **Method not found**: The method does not exist or is not available.
	ErrorCodeMethodNotFound ErrorCode = -32601
	// ErrorCodeInvalidParams: **Invalid params**: Invalid method parameter(s).
	ErrorCodeInvalidParams ErrorCode = -32602
	// ErrorCodeInternalError: **Internal error**: Internal JSON-RPC error.
	// Reserved for implementation-defined server errors.
	ErrorCodeInternalError ErrorCode = -32603
	// ErrorCodeRequestCanceled: **Request cancelled**: Execution of the method was aborted either due to a cancellation request from the caller or
	// because of resource constraints or shutdown.
	ErrorCodeRequestCanceled ErrorCode = -32800
	// ErrorCodeAuthenticationRequired: **Authentication required**: Authentication is required before this operation can be performed.
	ErrorCodeAuthenticationRequired ErrorCode = -32000
	// ErrorCodeResourceNotFound: **Resource not found**: A given resource, such as a file, was not found.
	ErrorCodeResourceNotFound ErrorCode = -32002
)

// ExtNotification: Allows the Agent to send an arbitrary notification that is not part of the ACP spec.
// Extension notifications provide a way to send one-way messages for custom functionality
// while maintaining protocol compatibility.
//
// See protocol docs: [Extensibility](https://agentclientprotocol.com/protocol/extensibility)
type ExtNotification any

// ExtRequest: Allows for sending an arbitrary request that is not part of the ACP spec.
// Extension methods provide a way to add custom functionality while maintaining
// protocol compatibility.
//
// See protocol docs: [Extensibility](https://agentclientprotocol.com/protocol/extensibility)
type ExtRequest any

// ExtResponse: Allows for sending an arbitrary response to an [`ExtRequest`] that is not part of the ACP spec.
// Extension methods provide a way to add custom functionality while maintaining
// protocol compatibility.
//
// See protocol docs: [Extensibility](https://agentclientprotocol.com/protocol/extensibility)
type ExtResponse any

// FileSystemCapabilities: File system capabilities that a client may support.
//
// See protocol docs: [FileSystem](https://agentclientprotocol.com/protocol/initialization#filesystem)
type FileSystemCapabilities struct {
	Meta          Meta `json:"_meta,omitzero"`
	ReadTextFile  bool `json:"readTextFile,omitempty"`
	WriteTextFile bool `json:"writeTextFile,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *FileSystemCapabilities) UnmarshalJSON(data []byte) error {
	type alias FileSystemCapabilities
	decoded := alias{}
	raw := struct {
		ReadTextFile  json.RawMessage `json:"readTextFile"`
		WriteTextFile json.RawMessage `json:"writeTextFile"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.ReadTextFile) > 0 {
		unmarshalDefault(raw.ReadTextFile, &decoded.ReadTextFile)
	}
	if len(raw.WriteTextFile) > 0 {
		unmarshalDefault(raw.WriteTextFile, &decoded.WriteTextFile)
	}
	*c = FileSystemCapabilities(decoded)
	return nil
}

// HTTPHeader: An HTTP header to set when making requests to the MCP server.
type HTTPHeader struct {
	Meta  Meta   `json:"_meta,omitzero"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ImageContent: An image provided to or from an LLM.
type ImageContent struct {
	Meta        Meta         `json:"_meta,omitzero"`
	Annotations *Annotations `json:"annotations,omitempty"`
	Data        string       `json:"data"`
	MIMEType    string       `json:"mimeType"`
	URI         *string      `json:"uri,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *ImageContent) UnmarshalJSON(data []byte) error {
	type alias ImageContent
	decoded := alias{}
	raw := struct {
		Annotations json.RawMessage `json:"annotations"`
		URI         json.RawMessage `json:"uri"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Annotations) > 0 {
		unmarshalDefault(raw.Annotations, &decoded.Annotations)
	}
	if len(raw.URI) > 0 {
		unmarshalDefault(raw.URI, &decoded.URI)
	}
	if err := requireJSONFieldsOnly(data, "data", "mimeType"); err != nil {
		return err
	}
	*c = ImageContent(decoded)
	return nil
}

// Implementation: Metadata about the implementation of the client or agent.
// Describes the name and version of an ACP implementation, with an optional
// title for UI representation.
type Implementation struct {
	Meta    Meta    `json:"_meta,omitzero"`
	Name    string  `json:"name"`
	Title   *string `json:"title,omitempty"`
	Version string  `json:"version"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (i *Implementation) UnmarshalJSON(data []byte) error {
	type alias Implementation
	decoded := alias{}
	raw := struct {
		Title json.RawMessage `json:"title"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Title) > 0 {
		unmarshalDefault(raw.Title, &decoded.Title)
	}
	if err := requireJSONFieldsOnly(data, "name", "version"); err != nil {
		return err
	}
	*i = Implementation(decoded)
	return nil
}

// InitializeRequest: Request parameters for the initialize method.
//
// Sent by the client to establish connection and negotiate capabilities.
//
// See protocol docs: [Initialization](https://agentclientprotocol.com/protocol/initialization)
type InitializeRequest struct {
	Meta               Meta                `json:"_meta,omitzero"`
	ClientCapabilities *ClientCapabilities `json:"clientCapabilities,omitempty"`
	ClientInfo         *Implementation     `json:"clientInfo,omitempty"`
	ProtocolVersion    ProtocolVersion     `json:"protocolVersion"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *InitializeRequest) UnmarshalJSON(data []byte) error {
	type alias InitializeRequest
	decoded := alias{}
	raw := struct {
		ClientCapabilities json.RawMessage `json:"clientCapabilities"`
		ClientInfo         json.RawMessage `json:"clientInfo"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.ClientCapabilities) > 0 {
		unmarshalDefault(raw.ClientCapabilities, &decoded.ClientCapabilities)
	}
	if len(raw.ClientInfo) > 0 {
		unmarshalDefault(raw.ClientInfo, &decoded.ClientInfo)
	}
	*r = InitializeRequest(decoded)
	return nil
}

// InitializeResponse: Response to the `initialize` method.
//
// Contains the negotiated protocol version and agent capabilities.
//
// See protocol docs: [Initialization](https://agentclientprotocol.com/protocol/initialization)
type InitializeResponse struct {
	Meta              Meta               `json:"_meta,omitzero"`
	AgentCapabilities *AgentCapabilities `json:"agentCapabilities,omitempty"`
	AgentInfo         *Implementation    `json:"agentInfo,omitempty"`
	AuthMethods       []AuthMethod       `json:"authMethods,omitempty"`
	ProtocolVersion   ProtocolVersion    `json:"protocolVersion"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *InitializeResponse) UnmarshalJSON(data []byte) error {
	type alias InitializeResponse
	decoded := alias{}
	raw := struct {
		AgentCapabilities json.RawMessage `json:"agentCapabilities"`
		AgentInfo         json.RawMessage `json:"agentInfo"`
		AuthMethods       json.RawMessage `json:"authMethods"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.AgentCapabilities) > 0 {
		unmarshalDefault(raw.AgentCapabilities, &decoded.AgentCapabilities)
	}
	if len(raw.AgentInfo) > 0 {
		unmarshalDefault(raw.AgentInfo, &decoded.AgentInfo)
	}
	if len(raw.AuthMethods) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.AuthMethods, &values); err == nil {
			decoded.AuthMethods = []AuthMethod{}
			for _, value := range values {
				var item AuthMethod
				if err := json.Unmarshal(value, &item); err == nil {
					decoded.AuthMethods = append(decoded.AuthMethods, item)
				}
			}
		}
	}
	*r = InitializeResponse(decoded)
	return nil
}

// IntegerPropertySchema: Schema for integer properties in an elicitation form.
type IntegerPropertySchema struct {
	Meta        Meta    `json:"_meta,omitzero"`
	Default     *int64  `json:"default,omitempty"`
	Description *string `json:"description,omitempty"`
	Maximum     *int64  `json:"maximum,omitempty"`
	Minimum     *int64  `json:"minimum,omitempty"`
	Title       *string `json:"title,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *IntegerPropertySchema) UnmarshalJSON(data []byte) error {
	type alias IntegerPropertySchema
	decoded := alias{}
	raw := struct {
		Default     json.RawMessage `json:"default"`
		Description json.RawMessage `json:"description"`
		Title       json.RawMessage `json:"title"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Default) > 0 {
		unmarshalDefault(raw.Default, &decoded.Default)
	}
	if len(raw.Description) > 0 {
		unmarshalDefault(raw.Description, &decoded.Description)
	}
	if len(raw.Title) > 0 {
		unmarshalDefault(raw.Title, &decoded.Title)
	}
	*s = IntegerPropertySchema(decoded)
	return nil
}

// KillTerminalRequest: Request to kill a terminal without releasing it.
type KillTerminalRequest struct {
	Meta       Meta       `json:"_meta,omitzero"`
	SessionID  SessionID  `json:"sessionId"`
	TerminalID TerminalID `json:"terminalId"`
}

// KillTerminalResponse: Response to `terminal/kill` method
type KillTerminalResponse struct {
	Meta Meta `json:"_meta,omitzero"`
}

// ListSessionsRequest: Request parameters for listing existing sessions.
//
// Only available if the Agent supports the `sessionCapabilities.list` capability.
type ListSessionsRequest struct {
	Meta   Meta    `json:"_meta,omitzero"`
	Cursor *string `json:"cursor,omitempty"`
	Cwd    *string `json:"cwd,omitempty"`
}

// ListSessionsResponse: Response from listing sessions.
type ListSessionsResponse struct {
	Meta       Meta          `json:"_meta,omitzero"`
	NextCursor *string       `json:"nextCursor,omitempty"`
	Sessions   []SessionInfo `json:"sessions"`
}

// LoadSessionRequest: Request parameters for loading an existing session.
//
// Only available if the Agent supports the `loadSession` capability.
//
// See protocol docs: [Loading Sessions](https://agentclientprotocol.com/protocol/session-setup#loading-sessions)
type LoadSessionRequest struct {
	Meta                  Meta        `json:"_meta,omitzero"`
	AdditionalDirectories []string    `json:"additionalDirectories,omitempty"`
	Cwd                   string      `json:"cwd"`
	MCPServers            []MCPServer `json:"mcpServers"`
	SessionID             SessionID   `json:"sessionId"`
}

// MarshalJSON implements json.Marshaler.
func (r LoadSessionRequest) MarshalJSON() ([]byte, error) {
	type alias LoadSessionRequest
	a := alias(r)
	if a.MCPServers == nil {
		a.MCPServers = []MCPServer{}
	}
	return json.Marshal(a)
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *LoadSessionRequest) UnmarshalJSON(data []byte) error {
	type alias LoadSessionRequest
	decoded := alias{}
	raw := struct {
		AdditionalDirectories json.RawMessage `json:"additionalDirectories"`
		MCPServers            json.RawMessage `json:"mcpServers"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.AdditionalDirectories) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.AdditionalDirectories, &values); err == nil {
			decoded.AdditionalDirectories = []string{}
			for _, value := range values {
				var item string
				if err := json.Unmarshal(value, &item); err == nil {
					decoded.AdditionalDirectories = append(decoded.AdditionalDirectories, item)
				}
			}
		}
	}
	if len(raw.MCPServers) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.MCPServers, &values); err == nil {
			decoded.MCPServers = []MCPServer{}
			for _, value := range values {
				var item MCPServer
				if err := json.Unmarshal(value, &item); err == nil {
					switch item.Type {
					case MCPServerTypeHTTP, MCPServerTypeSSE, "":
						decoded.MCPServers = append(decoded.MCPServers, item)
					}
				}
			}
		}
	}
	*r = LoadSessionRequest(decoded)
	return nil
}

// LoadSessionResponse: Response from loading an existing session.
type LoadSessionResponse struct {
	Meta          Meta                   `json:"_meta,omitzero"`
	ConfigOptions *[]SessionConfigOption `json:"configOptions,omitempty"`
	Modes         *SessionModeState      `json:"modes,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *LoadSessionResponse) UnmarshalJSON(data []byte) error {
	type alias LoadSessionResponse
	decoded := alias{}
	raw := struct {
		ConfigOptions json.RawMessage `json:"configOptions"`
		Modes         json.RawMessage `json:"modes"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.ConfigOptions) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.ConfigOptions, &values); err == nil && values != nil {
			items := []SessionConfigOption{}
			for _, value := range values {
				var item SessionConfigOption
				if err := json.Unmarshal(value, &item); err == nil {
					switch item.Type {
					case SessionConfigOptionTypeSelect, SessionConfigOptionTypeBoolean:
						items = append(items, item)
					}
				}
			}
			decoded.ConfigOptions = &items
		}
	}
	if len(raw.Modes) > 0 {
		unmarshalDefault(raw.Modes, &decoded.Modes)
	}
	*r = LoadSessionResponse(decoded)
	return nil
}

// LogoutCapabilities: Logout capabilities supported by the agent.
//
// Supplying `{}` means the agent supports the logout method.
type LogoutCapabilities struct {
	Meta Meta `json:"_meta,omitzero"`
}

// LogoutRequest: Request parameters for the logout method.
//
// Terminates the current authenticated session.
type LogoutRequest struct {
	Meta Meta `json:"_meta,omitzero"`
}

// LogoutResponse: Response to the `logout` method.
type LogoutResponse struct {
	Meta Meta `json:"_meta,omitzero"`
}

// MCPCapabilities: MCP capabilities supported by the agent
type MCPCapabilities struct {
	Meta Meta `json:"_meta,omitzero"`
	HTTP bool `json:"http,omitempty"`
	SSE  bool `json:"sse,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *MCPCapabilities) UnmarshalJSON(data []byte) error {
	type alias MCPCapabilities
	decoded := alias{}
	raw := struct {
		HTTP json.RawMessage `json:"http"`
		SSE  json.RawMessage `json:"sse"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.HTTP) > 0 {
		unmarshalDefault(raw.HTTP, &decoded.HTTP)
	}
	if len(raw.SSE) > 0 {
		unmarshalDefault(raw.SSE, &decoded.SSE)
	}
	*c = MCPCapabilities(decoded)
	return nil
}

// MCPServer: Configuration for connecting to an MCP (Model Context Protocol) server.
//
// MCP servers provide tools and context that the agent can use when
// processing prompts.
//
// See protocol docs: [MCP Servers](https://agentclientprotocol.com/protocol/session-setup#mcp-servers)
type MCPServer struct {
	Type    MCPServerType `json:"type,omitempty"`
	Meta    Meta          `json:"_meta,omitzero"`
	Args    []string      `json:"args,omitempty"`
	Command string        `json:"command,omitempty"`
	Env     []EnvVariable `json:"env,omitempty"`
	Headers []HTTPHeader  `json:"headers,omitempty"`
	Name    string        `json:"name"`
	URL     string        `json:"url,omitempty"`
}

// MCPServerType is the discriminator for MCPServer variants.
type MCPServerType string

const (
	MCPServerTypeHTTP MCPServerType = "http"
	MCPServerTypeSSE  MCPServerType = "sse"
)

// HTTPMCPServer creates an MCPServer variant: HTTP transport configuration
//
// Only available when the Agent capabilities indicate `mcp_capabilities.http` is `true`.
func HTTPMCPServer(name string, url string, headers []HTTPHeader) MCPServer {
	return MCPServer{
		Type:    MCPServerTypeHTTP,
		Headers: headers,
		Name:    name,
		URL:     url,
	}
}

// SSEMCPServer creates an MCPServer variant: SSE transport configuration
//
// Only available when the Agent capabilities indicate `mcp_capabilities.sse` is `true`.
func SSEMCPServer(name string, url string, headers []HTTPHeader) MCPServer {
	return MCPServer{
		Type:    MCPServerTypeSSE,
		Headers: headers,
		Name:    name,
		URL:     url,
	}
}

// StdioMCPServer creates an MCPServer variant: Stdio transport configuration
//
// All Agents MUST support this transport.
func StdioMCPServer(name string, command string, args []string, env []EnvVariable) MCPServer {
	return MCPServer{
		Args:    args,
		Command: command,
		Env:     env,
		Name:    name,
	}
}

// MarshalJSON implements json.Marshaler.
func (s MCPServer) MarshalJSON() ([]byte, error) {
	type alias MCPServer
	type wire struct {
		*alias
		Headers *[]HTTPHeader  `json:"headers,omitempty"`
		URL     *string        `json:"url,omitempty"`
		Args    *[]string      `json:"args,omitempty"`
		Command *string        `json:"command,omitempty"`
		Env     *[]EnvVariable `json:"env,omitempty"`
	}
	w := wire{alias: (*alias)(&s)}
	if len(s.Headers) > 0 {
		Headers := s.Headers
		w.Headers = &Headers
	}
	if !reflect.ValueOf(s.URL).IsZero() {
		URL := s.URL
		w.URL = &URL
	}
	if len(s.Args) > 0 {
		Args := s.Args
		w.Args = &Args
	}
	if !reflect.ValueOf(s.Command).IsZero() {
		Command := s.Command
		w.Command = &Command
	}
	if len(s.Env) > 0 {
		Env := s.Env
		w.Env = &Env
	}
	switch s.Type {
	case MCPServerTypeHTTP:
		Headers := s.Headers
		if Headers == nil {
			Headers = []HTTPHeader{}
		}
		w.Headers = &Headers
		URL := s.URL
		w.URL = &URL
	case MCPServerTypeSSE:
		Headers := s.Headers
		if Headers == nil {
			Headers = []HTTPHeader{}
		}
		w.Headers = &Headers
		URL := s.URL
		w.URL = &URL
	case "":
		Args := s.Args
		if Args == nil {
			Args = []string{}
		}
		w.Args = &Args
		Command := s.Command
		w.Command = &Command
		Env := s.Env
		if Env == nil {
			Env = []EnvVariable{}
		}
		w.Env = &Env
	}
	return json.Marshal(w)
}

// MCPServerHTTP: HTTP transport configuration for MCP.
type MCPServerHTTP struct {
	Meta    Meta         `json:"_meta,omitzero"`
	Headers []HTTPHeader `json:"headers"`
	Name    string       `json:"name"`
	URL     string       `json:"url"`
}

// MarshalJSON implements json.Marshaler.
func (h MCPServerHTTP) MarshalJSON() ([]byte, error) {
	type alias MCPServerHTTP
	a := alias(h)
	if a.Headers == nil {
		a.Headers = []HTTPHeader{}
	}
	return json.Marshal(a)
}

// MCPServerSSE: SSE transport configuration for MCP.
type MCPServerSSE struct {
	Meta    Meta         `json:"_meta,omitzero"`
	Headers []HTTPHeader `json:"headers"`
	Name    string       `json:"name"`
	URL     string       `json:"url"`
}

// MarshalJSON implements json.Marshaler.
func (s MCPServerSSE) MarshalJSON() ([]byte, error) {
	type alias MCPServerSSE
	a := alias(s)
	if a.Headers == nil {
		a.Headers = []HTTPHeader{}
	}
	return json.Marshal(a)
}

// MCPServerStdio: Stdio transport configuration for MCP.
type MCPServerStdio struct {
	Meta    Meta          `json:"_meta,omitzero"`
	Args    []string      `json:"args"`
	Command string        `json:"command"`
	Env     []EnvVariable `json:"env"`
	Name    string        `json:"name"`
}

// MarshalJSON implements json.Marshaler.
func (s MCPServerStdio) MarshalJSON() ([]byte, error) {
	type alias MCPServerStdio
	a := alias(s)
	if a.Args == nil {
		a.Args = []string{}
	}
	if a.Env == nil {
		a.Env = []EnvVariable{}
	}
	return json.Marshal(a)
}

// MessageID: Unique identifier for a message within a session.
type MessageID string

// MultiSelectItems: Items for a multi-select (array) property schema.
type MultiSelectItems struct {
	Type   MultiSelectItemsType       `json:"type,omitempty"`
	Meta   Meta                       `json:"_meta,omitzero"`
	Fields map[string]json.RawMessage `json:"-"`
	AnyOf  []EnumOption               `json:"anyOf,omitempty"`
	Enum   []string                   `json:"enum,omitempty"`
}

// MultiSelectItemsType is the discriminator for MultiSelectItems variants.
type MultiSelectItemsType string

const (
	MultiSelectItemsTypeString MultiSelectItemsType = "string"
)

// NewStringMultiSelectItems creates MultiSelectItems with plain string values.
func NewStringMultiSelectItems(enum []string) MultiSelectItems {
	return MultiSelectItems{
		Type: MultiSelectItemsTypeString,
		Enum: enum,
	}
}

// OtherMultiSelectItems creates custom or future typed MultiSelectItems.
func OtherMultiSelectItems(typeName string, fields map[string]json.RawMessage) MultiSelectItems {
	if typeName == "" {
		copied := make(map[string]json.RawMessage, len(fields)+1)
		maps.Copy(copied, fields)
		copied["type"] = json.RawMessage(`""`)
		fields = copied
	}
	return MultiSelectItems{
		Type:   MultiSelectItemsType(typeName),
		Fields: fields,
	}
}

// NewTitledMultiSelectItems creates MultiSelectItems with human-readable labels.
func NewTitledMultiSelectItems(anyOf []EnumOption) MultiSelectItems {
	return MultiSelectItems{
		AnyOf: anyOf,
	}
}

// MarshalJSON implements json.Marshaler.
func (i MultiSelectItems) MarshalJSON() ([]byte, error) {
	type alias MultiSelectItems
	type wire struct {
		*alias
		Enum  *[]string     `json:"enum,omitempty"`
		AnyOf *[]EnumOption `json:"anyOf,omitempty"`
	}
	w := wire{alias: (*alias)(&i)}
	if len(i.Enum) > 0 {
		Enum := i.Enum
		w.Enum = &Enum
	}
	if len(i.AnyOf) > 0 {
		AnyOf := i.AnyOf
		w.AnyOf = &AnyOf
	}
	switch i.Type {
	case MultiSelectItemsTypeString:
		Enum := i.Enum
		if Enum == nil {
			Enum = []string{}
		}
		w.Enum = &Enum
	case "":
		if _, isOther := i.Fields["type"]; !isOther {
			AnyOf := i.AnyOf
			if AnyOf == nil {
				AnyOf = []EnumOption{}
			}
			w.AnyOf = &AnyOf
		}
	}
	return marshalJSONWithFields(w, i.Fields)
}

// UnmarshalJSON implements json.Unmarshaler.
func (i *MultiSelectItems) UnmarshalJSON(data []byte) error {
	fields, err := requireJSONKeys(data)
	if err != nil {
		return err
	}
	rawType, hasType := fields["type"]
	if hasType {
		if jsonValueIsNull(rawType) {
			return errors.New("acp: multi-select item type must be a string")
		}
		var itemType MultiSelectItemsType
		if err := json.Unmarshal(rawType, &itemType); err != nil {
			return err
		}
		if itemType == MultiSelectItemsTypeString {
			var value StringMultiSelectItems
			if err := json.Unmarshal(data, &value); err != nil {
				return err
			}
			*i = MultiSelectItems{Type: itemType, Meta: value.Meta, Enum: value.Enum}
			return nil
		}
		if itemType != "" {
			delete(fields, "type")
		}
		*i = MultiSelectItems{Type: itemType, Fields: fields}
		return nil
	}

	var value TitledMultiSelectItems
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*i = MultiSelectItems{Meta: value.Meta, AnyOf: value.AnyOf}
	return nil
}

// MultiSelectPropertySchema: Schema for multi-select (array) properties in an elicitation form.
type MultiSelectPropertySchema struct {
	Meta        Meta             `json:"_meta,omitzero"`
	Default     *[]string        `json:"default,omitempty"`
	Description *string          `json:"description,omitempty"`
	Items       MultiSelectItems `json:"items"`
	MaxItems    *uint64          `json:"maxItems,omitempty"`
	MinItems    *uint64          `json:"minItems,omitempty"`
	Title       *string          `json:"title,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *MultiSelectPropertySchema) UnmarshalJSON(data []byte) error {
	type alias MultiSelectPropertySchema
	decoded := alias{}
	raw := struct {
		Default     json.RawMessage `json:"default"`
		Description json.RawMessage `json:"description"`
		Title       json.RawMessage `json:"title"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Default) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.Default, &values); err == nil && values != nil {
			items := []string{}
			for _, value := range values {
				var item string
				if err := json.Unmarshal(value, &item); err == nil {
					items = append(items, item)
				}
			}
			decoded.Default = &items
		}
	}
	if len(raw.Description) > 0 {
		unmarshalDefault(raw.Description, &decoded.Description)
	}
	if len(raw.Title) > 0 {
		unmarshalDefault(raw.Title, &decoded.Title)
	}
	if err := requireJSONFieldsOnly(data, "items"); err != nil {
		return err
	}
	*s = MultiSelectPropertySchema(decoded)
	return nil
}

// NewSessionRequest: Request parameters for creating a new session.
//
// See protocol docs: [Creating a Session](https://agentclientprotocol.com/protocol/session-setup#creating-a-session)
type NewSessionRequest struct {
	Meta                  Meta        `json:"_meta,omitzero"`
	AdditionalDirectories []string    `json:"additionalDirectories,omitempty"`
	Cwd                   string      `json:"cwd"`
	MCPServers            []MCPServer `json:"mcpServers"`
}

// MarshalJSON implements json.Marshaler.
func (r NewSessionRequest) MarshalJSON() ([]byte, error) {
	type alias NewSessionRequest
	a := alias(r)
	if a.MCPServers == nil {
		a.MCPServers = []MCPServer{}
	}
	return json.Marshal(a)
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *NewSessionRequest) UnmarshalJSON(data []byte) error {
	type alias NewSessionRequest
	decoded := alias{}
	raw := struct {
		AdditionalDirectories json.RawMessage `json:"additionalDirectories"`
		MCPServers            json.RawMessage `json:"mcpServers"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.AdditionalDirectories) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.AdditionalDirectories, &values); err == nil {
			decoded.AdditionalDirectories = []string{}
			for _, value := range values {
				var item string
				if err := json.Unmarshal(value, &item); err == nil {
					decoded.AdditionalDirectories = append(decoded.AdditionalDirectories, item)
				}
			}
		}
	}
	if len(raw.MCPServers) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.MCPServers, &values); err == nil {
			decoded.MCPServers = []MCPServer{}
			for _, value := range values {
				var item MCPServer
				if err := json.Unmarshal(value, &item); err == nil {
					switch item.Type {
					case MCPServerTypeHTTP, MCPServerTypeSSE, "":
						decoded.MCPServers = append(decoded.MCPServers, item)
					}
				}
			}
		}
	}
	*r = NewSessionRequest(decoded)
	return nil
}

// NewSessionResponse: Response from creating a new session.
//
// See protocol docs: [Creating a Session](https://agentclientprotocol.com/protocol/session-setup#creating-a-session)
type NewSessionResponse struct {
	Meta          Meta                   `json:"_meta,omitzero"`
	ConfigOptions *[]SessionConfigOption `json:"configOptions,omitempty"`
	Modes         *SessionModeState      `json:"modes,omitempty"`
	SessionID     SessionID              `json:"sessionId"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *NewSessionResponse) UnmarshalJSON(data []byte) error {
	type alias NewSessionResponse
	decoded := alias{}
	raw := struct {
		ConfigOptions json.RawMessage `json:"configOptions"`
		Modes         json.RawMessage `json:"modes"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.ConfigOptions) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.ConfigOptions, &values); err == nil && values != nil {
			items := []SessionConfigOption{}
			for _, value := range values {
				var item SessionConfigOption
				if err := json.Unmarshal(value, &item); err == nil {
					switch item.Type {
					case SessionConfigOptionTypeSelect, SessionConfigOptionTypeBoolean:
						items = append(items, item)
					}
				}
			}
			decoded.ConfigOptions = &items
		}
	}
	if len(raw.Modes) > 0 {
		unmarshalDefault(raw.Modes, &decoded.Modes)
	}
	*r = NewSessionResponse(decoded)
	return nil
}

// NumberPropertySchema: Schema for number (floating-point) properties in an elicitation form.
type NumberPropertySchema struct {
	Meta        Meta     `json:"_meta,omitzero"`
	Default     *float64 `json:"default,omitempty"`
	Description *string  `json:"description,omitempty"`
	Maximum     *float64 `json:"maximum,omitempty"`
	Minimum     *float64 `json:"minimum,omitempty"`
	Title       *string  `json:"title,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *NumberPropertySchema) UnmarshalJSON(data []byte) error {
	type alias NumberPropertySchema
	decoded := alias{}
	raw := struct {
		Default     json.RawMessage `json:"default"`
		Description json.RawMessage `json:"description"`
		Title       json.RawMessage `json:"title"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Default) > 0 {
		unmarshalDefault(raw.Default, &decoded.Default)
	}
	if len(raw.Description) > 0 {
		unmarshalDefault(raw.Description, &decoded.Description)
	}
	if len(raw.Title) > 0 {
		unmarshalDefault(raw.Title, &decoded.Title)
	}
	*s = NumberPropertySchema(decoded)
	return nil
}

// PermissionOption: An option presented to the user when requesting permission.
type PermissionOption struct {
	Meta     Meta                 `json:"_meta,omitzero"`
	Kind     PermissionOptionKind `json:"kind"`
	Name     string               `json:"name"`
	OptionID PermissionOptionID   `json:"optionId"`
}

// PermissionOptionID: Unique identifier for a permission option.
type PermissionOptionID string

// PermissionOptionKind: The type of permission option being presented to the user.
//
// Helps clients choose appropriate icons and UI treatment.
type PermissionOptionKind string

const (
	// PermissionOptionKindAllowOnce: Allow this operation only this time.
	PermissionOptionKindAllowOnce PermissionOptionKind = "allow_once"
	// PermissionOptionKindAllowAlways: Allow this operation and remember the choice.
	PermissionOptionKindAllowAlways PermissionOptionKind = "allow_always"
	// PermissionOptionKindRejectOnce: Reject this operation only this time.
	PermissionOptionKindRejectOnce PermissionOptionKind = "reject_once"
	// PermissionOptionKindRejectAlways: Reject this operation and remember the choice.
	PermissionOptionKindRejectAlways PermissionOptionKind = "reject_always"
)

// Plan: An execution plan for accomplishing complex tasks.
//
// Plans consist of multiple entries representing individual tasks or goals.
// Agents report plans to clients to provide visibility into their execution strategy.
// Plans can evolve during execution as the agent discovers new requirements or completes tasks.
//
// See protocol docs: [Agent Plan](https://agentclientprotocol.com/protocol/agent-plan)
type Plan struct {
	Meta    Meta        `json:"_meta,omitzero"`
	Entries []PlanEntry `json:"entries"`
}

// MarshalJSON implements json.Marshaler.
func (p Plan) MarshalJSON() ([]byte, error) {
	type alias Plan
	a := alias(p)
	if a.Entries == nil {
		a.Entries = []PlanEntry{}
	}
	return json.Marshal(a)
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *Plan) UnmarshalJSON(data []byte) error {
	type alias Plan
	decoded := alias{}
	raw := struct {
		Entries json.RawMessage `json:"entries"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Entries) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.Entries, &values); err == nil {
			decoded.Entries = []PlanEntry{}
			for _, value := range values {
				var item PlanEntry
				if err := json.Unmarshal(value, &item); err == nil {
					decoded.Entries = append(decoded.Entries, item)
				}
			}
		}
	}
	if err := requireJSONKeysOnly(data, "entries"); err != nil {
		return err
	}
	*p = Plan(decoded)
	return nil
}

// PlanEntry: A single entry in the execution plan.
//
// Represents a task or goal that the assistant intends to accomplish
// as part of fulfilling the user's request.
// See protocol docs: [Plan Entries](https://agentclientprotocol.com/protocol/agent-plan#plan-entries)
type PlanEntry struct {
	Meta     Meta              `json:"_meta,omitzero"`
	Content  string            `json:"content"`
	Priority PlanEntryPriority `json:"priority"`
	Status   PlanEntryStatus   `json:"status"`
}

// PlanEntryPriority: Priority levels for plan entries.
//
// Used to indicate the relative importance or urgency of different
// tasks in the execution plan.
// See protocol docs: [Plan Entries](https://agentclientprotocol.com/protocol/agent-plan#plan-entries)
type PlanEntryPriority string

const (
	// PlanEntryPriorityHigh: High priority task - critical to the overall goal.
	PlanEntryPriorityHigh PlanEntryPriority = "high"
	// PlanEntryPriorityMedium: Medium priority task - important but not critical.
	PlanEntryPriorityMedium PlanEntryPriority = "medium"
	// PlanEntryPriorityLow: Low priority task - nice to have but not essential.
	PlanEntryPriorityLow PlanEntryPriority = "low"
)

// PlanEntryStatus: Status of a plan entry in the execution flow.
//
// Tracks the lifecycle of each task from planning through completion.
// See protocol docs: [Plan Entries](https://agentclientprotocol.com/protocol/agent-plan#plan-entries)
type PlanEntryStatus string

const (
	// PlanEntryStatusPending: The task has not started yet.
	PlanEntryStatusPending PlanEntryStatus = "pending"
	// PlanEntryStatusInProgress: The task is currently being worked on.
	PlanEntryStatusInProgress PlanEntryStatus = "in_progress"
	// PlanEntryStatusCompleted: The task has been successfully completed.
	PlanEntryStatusCompleted PlanEntryStatus = "completed"
)

// PromptCapabilities: Prompt capabilities supported by the agent in `session/prompt` requests.
//
// Baseline agent functionality requires support for [`ContentBlock::Text`]
// and [`ContentBlock::ResourceLink`] in prompt requests.
//
// Other variants must be explicitly opted in to.
// Capabilities for different types of content in prompt requests.
//
// Indicates which content types beyond the baseline (text and resource links)
// the agent can process.
//
// See protocol docs: [Prompt Capabilities](https://agentclientprotocol.com/protocol/initialization#prompt-capabilities)
type PromptCapabilities struct {
	Meta            Meta `json:"_meta,omitzero"`
	Audio           bool `json:"audio,omitempty"`
	EmbeddedContext bool `json:"embeddedContext,omitempty"`
	Image           bool `json:"image,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *PromptCapabilities) UnmarshalJSON(data []byte) error {
	type alias PromptCapabilities
	decoded := alias{}
	raw := struct {
		Audio           json.RawMessage `json:"audio"`
		EmbeddedContext json.RawMessage `json:"embeddedContext"`
		Image           json.RawMessage `json:"image"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Audio) > 0 {
		unmarshalDefault(raw.Audio, &decoded.Audio)
	}
	if len(raw.EmbeddedContext) > 0 {
		unmarshalDefault(raw.EmbeddedContext, &decoded.EmbeddedContext)
	}
	if len(raw.Image) > 0 {
		unmarshalDefault(raw.Image, &decoded.Image)
	}
	*c = PromptCapabilities(decoded)
	return nil
}

// PromptRequest: Request parameters for sending a user prompt to the agent.
//
// Contains the user's message and any additional context.
//
// See protocol docs: [User Message](https://agentclientprotocol.com/protocol/prompt-turn#1-user-message)
type PromptRequest struct {
	Meta      Meta           `json:"_meta,omitzero"`
	Prompt    []ContentBlock `json:"prompt"`
	SessionID SessionID      `json:"sessionId"`
}

// MarshalJSON implements json.Marshaler.
func (r PromptRequest) MarshalJSON() ([]byte, error) {
	type alias PromptRequest
	a := alias(r)
	if a.Prompt == nil {
		a.Prompt = []ContentBlock{}
	}
	return json.Marshal(a)
}

// PromptResponse: Response from processing a user prompt.
//
// See protocol docs: [Check for Completion](https://agentclientprotocol.com/protocol/prompt-turn#4-check-for-completion)
type PromptResponse struct {
	Meta       Meta       `json:"_meta,omitzero"`
	StopReason StopReason `json:"stopReason"`
	// Usage is unstable but already emitted by agents; retaining it prevents
	// clients from silently losing per-turn token accounting.
	Usage *Usage `json:"usage,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *PromptResponse) UnmarshalJSON(data []byte) error {
	type alias PromptResponse
	decoded := alias{}
	raw := struct {
		Usage json.RawMessage `json:"usage"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Usage) > 0 {
		// The SDK schema defaults malformed optional usage so an unstable
		// accounting extension cannot invalidate an otherwise completed turn.
		unmarshalDefault(raw.Usage, &decoded.Usage)
	}
	// Direct decoders bypass the RPC envelope's required-field validation.
	if err := requireJSONFieldsOnly(data, "stopReason"); err != nil {
		return err
	}
	*r = PromptResponse(decoded)
	return nil
}

// ProtocolVersion: Protocol version identifier.
//
// This version is only bumped for breaking changes.
// Non-breaking changes should be introduced via capabilities.
type ProtocolVersion uint16

// ReadTextFileRequest: Request to read content from a text file.
//
// Only available if the client supports the `fs.readTextFile` capability.
type ReadTextFileRequest struct {
	Meta      Meta      `json:"_meta,omitzero"`
	Limit     *uint32   `json:"limit,omitempty"`
	Line      *uint32   `json:"line,omitempty"`
	Path      string    `json:"path"`
	SessionID SessionID `json:"sessionId"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *ReadTextFileRequest) UnmarshalJSON(data []byte) error {
	type alias ReadTextFileRequest
	decoded := alias{}
	raw := struct {
		Limit json.RawMessage `json:"limit"`
		Line  json.RawMessage `json:"line"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Limit) > 0 {
		unmarshalDefault(raw.Limit, &decoded.Limit)
	}
	if len(raw.Line) > 0 {
		unmarshalDefault(raw.Line, &decoded.Line)
	}
	*r = ReadTextFileRequest(decoded)
	return nil
}

// ReadTextFileResponse: Response containing the contents of a text file.
type ReadTextFileResponse struct {
	Meta    Meta   `json:"_meta,omitzero"`
	Content string `json:"content"`
}

// ReleaseTerminalRequest: Request to release a terminal and free its resources.
type ReleaseTerminalRequest struct {
	Meta       Meta       `json:"_meta,omitzero"`
	SessionID  SessionID  `json:"sessionId"`
	TerminalID TerminalID `json:"terminalId"`
}

// ReleaseTerminalResponse: Response to terminal/release method
type ReleaseTerminalResponse struct {
	Meta Meta `json:"_meta,omitzero"`
}

// RequestID: JSON RPC Request Id
//
// An identifier established by the Client that MUST contain a String, Number, or NULL value if included. If it is not included it is assumed to be a notification. The value SHOULD normally not be Null \[1\] and Numbers SHOULD NOT contain fractional parts \[2\]
//
// The Server MUST reply with the same value in the Response object if included. This member is used to correlate the context between the two objects.
//
// \[1\] The use of Null as a value for the id member in a Request object is discouraged, because this specification uses a value of Null for Responses with an unknown id. Also, because JSON-RPC 1.0 uses an id value of Null for Notifications this could cause confusion in handling.
//
// \[2\] Fractional parts may be problematic, since many decimal fractions cannot be represented exactly as binary fractions.
type RequestID = json.RawMessage

// RequestPermissionOutcome: The outcome of a permission request.
type RequestPermissionOutcome struct {
	Outcome  RequestPermissionOutcomeType `json:"outcome"`
	Meta     Meta                         `json:"_meta,omitzero"`
	OptionID PermissionOptionID           `json:"optionId,omitempty"`
}

// RequestPermissionOutcomeType is the discriminator for RequestPermissionOutcome variants.
type RequestPermissionOutcomeType string

const (
	RequestPermissionOutcomeTypeCanceled RequestPermissionOutcomeType = "cancelled"
	RequestPermissionOutcomeTypeSelected RequestPermissionOutcomeType = "selected"
)

// CanceledRequestPermissionOutcome creates an RequestPermissionOutcome variant: The prompt turn was cancelled before the user responded.
//
// When a client sends a `session/cancel` notification to cancel an ongoing
// prompt turn, it MUST respond to all pending `session/request_permission`
// requests with this `Cancelled` outcome.
//
// See protocol docs: [Cancellation](https://agentclientprotocol.com/protocol/prompt-turn#cancellation)
func CanceledRequestPermissionOutcome() RequestPermissionOutcome {
	return RequestPermissionOutcome{
		Outcome: RequestPermissionOutcomeTypeCanceled,
	}
}

// SelectedRequestPermissionOutcome creates an RequestPermissionOutcome variant: The user selected one of the provided options.
func SelectedRequestPermissionOutcome(optionID PermissionOptionID) RequestPermissionOutcome {
	return RequestPermissionOutcome{
		Outcome:  RequestPermissionOutcomeTypeSelected,
		OptionID: optionID,
	}
}

// MarshalJSON implements json.Marshaler.
func (o RequestPermissionOutcome) MarshalJSON() ([]byte, error) {
	type alias RequestPermissionOutcome
	type wire struct {
		*alias
		OptionID *PermissionOptionID `json:"optionId,omitempty"`
	}
	w := wire{alias: (*alias)(&o)}
	if !reflect.ValueOf(o.OptionID).IsZero() {
		OptionID := o.OptionID
		w.OptionID = &OptionID
	}
	switch o.Outcome {
	case RequestPermissionOutcomeTypeSelected:
		OptionID := o.OptionID
		w.OptionID = &OptionID
	case RequestPermissionOutcomeTypeCanceled:
	}
	return json.Marshal(w)
}

// RequestPermissionRequest: Request for user permission to execute a tool call.
//
// Sent when the agent needs authorization before performing a sensitive operation.
//
// See protocol docs: [Requesting Permission](https://agentclientprotocol.com/protocol/tool-calls#requesting-permission)
type RequestPermissionRequest struct {
	Meta      Meta               `json:"_meta,omitzero"`
	Options   []PermissionOption `json:"options"`
	SessionID SessionID          `json:"sessionId"`
	ToolCall  ToolCallUpdate     `json:"toolCall"`
}

// MarshalJSON implements json.Marshaler.
func (r RequestPermissionRequest) MarshalJSON() ([]byte, error) {
	type alias RequestPermissionRequest
	a := alias(r)
	if a.Options == nil {
		a.Options = []PermissionOption{}
	}
	return json.Marshal(a)
}

// RequestPermissionResponse: Response to a permission request.
type RequestPermissionResponse struct {
	Meta    Meta                     `json:"_meta,omitzero"`
	Outcome RequestPermissionOutcome `json:"outcome"`
}

// ResourceLink: A resource that the server is capable of reading, included in a prompt or tool call result.
type ResourceLink struct {
	Meta        Meta         `json:"_meta,omitzero"`
	Annotations *Annotations `json:"annotations,omitempty"`
	Description *string      `json:"description,omitempty"`
	MIMEType    *string      `json:"mimeType,omitempty"`
	Name        string       `json:"name"`
	Size        *int64       `json:"size,omitempty"`
	Title       *string      `json:"title,omitempty"`
	URI         string       `json:"uri"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (l *ResourceLink) UnmarshalJSON(data []byte) error {
	type alias ResourceLink
	decoded := alias{}
	raw := struct {
		Annotations json.RawMessage `json:"annotations"`
		Description json.RawMessage `json:"description"`
		MIMEType    json.RawMessage `json:"mimeType"`
		Size        json.RawMessage `json:"size"`
		Title       json.RawMessage `json:"title"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Annotations) > 0 {
		unmarshalDefault(raw.Annotations, &decoded.Annotations)
	}
	if len(raw.Description) > 0 {
		unmarshalDefault(raw.Description, &decoded.Description)
	}
	if len(raw.MIMEType) > 0 {
		unmarshalDefault(raw.MIMEType, &decoded.MIMEType)
	}
	if len(raw.Size) > 0 {
		unmarshalDefault(raw.Size, &decoded.Size)
	}
	if len(raw.Title) > 0 {
		unmarshalDefault(raw.Title, &decoded.Title)
	}
	if err := requireJSONFieldsOnly(data, "name", "uri"); err != nil {
		return err
	}
	*l = ResourceLink(decoded)
	return nil
}

// ResumeSessionRequest: Request parameters for resuming an existing session.
//
// Resumes an existing session without returning previous messages (unlike `session/load`).
// This is useful for agents that can resume sessions but don't implement full session loading.
//
// Only available if the Agent supports the `sessionCapabilities.resume` capability.
type ResumeSessionRequest struct {
	Meta                  Meta        `json:"_meta,omitzero"`
	AdditionalDirectories []string    `json:"additionalDirectories,omitempty"`
	Cwd                   string      `json:"cwd"`
	MCPServers            []MCPServer `json:"mcpServers,omitempty"`
	SessionID             SessionID   `json:"sessionId"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *ResumeSessionRequest) UnmarshalJSON(data []byte) error {
	type alias ResumeSessionRequest
	decoded := alias{}
	raw := struct {
		AdditionalDirectories json.RawMessage `json:"additionalDirectories"`
		MCPServers            json.RawMessage `json:"mcpServers"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.AdditionalDirectories) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.AdditionalDirectories, &values); err == nil {
			decoded.AdditionalDirectories = []string{}
			for _, value := range values {
				var item string
				if err := json.Unmarshal(value, &item); err == nil {
					decoded.AdditionalDirectories = append(decoded.AdditionalDirectories, item)
				}
			}
		}
	}
	if len(raw.MCPServers) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.MCPServers, &values); err == nil {
			decoded.MCPServers = []MCPServer{}
			for _, value := range values {
				var item MCPServer
				if err := json.Unmarshal(value, &item); err == nil {
					switch item.Type {
					case MCPServerTypeHTTP, MCPServerTypeSSE, "":
						decoded.MCPServers = append(decoded.MCPServers, item)
					}
				}
			}
		}
	}
	*r = ResumeSessionRequest(decoded)
	return nil
}

// ResumeSessionResponse: Response from resuming an existing session.
type ResumeSessionResponse struct {
	Meta          Meta                   `json:"_meta,omitzero"`
	ConfigOptions *[]SessionConfigOption `json:"configOptions,omitempty"`
	Modes         *SessionModeState      `json:"modes,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *ResumeSessionResponse) UnmarshalJSON(data []byte) error {
	type alias ResumeSessionResponse
	decoded := alias{}
	raw := struct {
		ConfigOptions json.RawMessage `json:"configOptions"`
		Modes         json.RawMessage `json:"modes"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.ConfigOptions) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.ConfigOptions, &values); err == nil && values != nil {
			items := []SessionConfigOption{}
			for _, value := range values {
				var item SessionConfigOption
				if err := json.Unmarshal(value, &item); err == nil {
					switch item.Type {
					case SessionConfigOptionTypeSelect, SessionConfigOptionTypeBoolean:
						items = append(items, item)
					}
				}
			}
			decoded.ConfigOptions = &items
		}
	}
	if len(raw.Modes) > 0 {
		unmarshalDefault(raw.Modes, &decoded.Modes)
	}
	*r = ResumeSessionResponse(decoded)
	return nil
}

// Role: The sender or recipient of messages and data in a conversation.
type Role string

const (
	// RoleAssistant: The assistant side of a conversation.
	RoleAssistant Role = "assistant"
	// RoleUser: The user side of a conversation.
	RoleUser Role = "user"
)

// SelectedPermissionOutcome: The user selected one of the provided options.
type SelectedPermissionOutcome struct {
	Meta     Meta               `json:"_meta,omitzero"`
	OptionID PermissionOptionID `json:"optionId"`
}

// SessionAdditionalDirectoriesCapabilities: Capabilities for additional session directories support.
//
// Supplying `{}` means the agent supports the `additionalDirectories` field on
// supported session lifecycle requests. Agents that also support
// `session/list` may return `SessionInfo.additionalDirectories` to report the
// complete ordered additional-root list associated with a listed session.
type SessionAdditionalDirectoriesCapabilities struct {
	Meta Meta `json:"_meta,omitzero"`
}

// SessionCapabilities: Session capabilities supported by the agent.
//
// As a baseline, all Agents **MUST** support `session/new`, `session/prompt`, `session/cancel`, and `session/update`.
//
// Optionally, they **MAY** support other session methods and notifications by specifying additional capabilities.
//
// Note: `session/load` is still handled by the top-level `load_session` capability. This will be unified in future versions of the protocol.
//
// See protocol docs: [Session Capabilities](https://agentclientprotocol.com/protocol/initialization#session-capabilities)
type SessionCapabilities struct {
	Meta                  Meta                                      `json:"_meta,omitzero"`
	AdditionalDirectories *SessionAdditionalDirectoriesCapabilities `json:"additionalDirectories,omitempty"`
	Close                 *SessionCloseCapabilities                 `json:"close,omitempty"`
	Delete                *SessionDeleteCapabilities                `json:"delete,omitempty"`
	List                  *SessionListCapabilities                  `json:"list,omitempty"`
	Resume                *SessionResumeCapabilities                `json:"resume,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *SessionCapabilities) UnmarshalJSON(data []byte) error {
	type alias SessionCapabilities
	decoded := alias{}
	raw := struct {
		AdditionalDirectories json.RawMessage `json:"additionalDirectories"`
		Close                 json.RawMessage `json:"close"`
		Delete                json.RawMessage `json:"delete"`
		List                  json.RawMessage `json:"list"`
		Resume                json.RawMessage `json:"resume"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.AdditionalDirectories) > 0 {
		unmarshalDefault(raw.AdditionalDirectories, &decoded.AdditionalDirectories)
	}
	if len(raw.Close) > 0 {
		unmarshalDefault(raw.Close, &decoded.Close)
	}
	if len(raw.Delete) > 0 {
		unmarshalDefault(raw.Delete, &decoded.Delete)
	}
	if len(raw.List) > 0 {
		unmarshalDefault(raw.List, &decoded.List)
	}
	if len(raw.Resume) > 0 {
		unmarshalDefault(raw.Resume, &decoded.Resume)
	}
	*c = SessionCapabilities(decoded)
	return nil
}

// SessionCloseCapabilities: Capabilities for the `session/close` method.
//
// Supplying `{}` means the agent supports closing sessions.
type SessionCloseCapabilities struct {
	Meta Meta `json:"_meta,omitzero"`
}

// SessionConfigBoolean: A boolean on/off toggle session configuration option payload.
type SessionConfigBoolean struct {
	CurrentValue bool `json:"currentValue"`
}

// SessionConfigGroupID: Unique identifier for a session configuration option value group.
type SessionConfigGroupID string

// SessionConfigID: Unique identifier for a session configuration option.
type SessionConfigID string

// SessionConfigOption: A session configuration option selector and its current state.
type SessionConfigOption struct {
	Type         SessionConfigOptionType      `json:"type"`
	Meta         Meta                         `json:"_meta,omitzero"`
	Category     *SessionConfigOptionCategory `json:"category,omitempty"`
	CurrentValue any                          `json:"currentValue"`
	Description  *string                      `json:"description,omitempty"`
	ID           SessionConfigID              `json:"id"`
	Name         string                       `json:"name"`
	Options      SessionConfigSelectOptions   `json:"options,omitzero"`
}

// SessionConfigOptionType is the discriminator for SessionConfigOption variants.
type SessionConfigOptionType string

const (
	SessionConfigOptionTypeSelect  SessionConfigOptionType = "select"
	SessionConfigOptionTypeBoolean SessionConfigOptionType = "boolean"
)

// SelectSessionConfigOption creates an SessionConfigOption variant: Single-value selector (dropdown).
func SelectSessionConfigOption(id SessionConfigID, name string, currentValue SessionConfigValueID, options SessionConfigSelectOptions) SessionConfigOption {
	if options.Ungrouped == nil && options.Groups == nil {
		flat := UngroupedSessionConfigSelectOptions{}
		options = SessionConfigSelectOptions{Ungrouped: &flat}
	}
	return SessionConfigOption{
		Type:         SessionConfigOptionTypeSelect,
		CurrentValue: currentValue,
		ID:           id,
		Name:         name,
		Options:      options,
	}
}

// BooleanSessionConfigOption creates an SessionConfigOption variant: Boolean on/off toggle.
func BooleanSessionConfigOption(id SessionConfigID, name string, currentValue bool) SessionConfigOption {
	return SessionConfigOption{
		Type:         SessionConfigOptionTypeBoolean,
		CurrentValue: currentValue,
		ID:           id,
		Name:         name,
	}
}

// MarshalJSON implements json.Marshaler.
func (o SessionConfigOption) MarshalJSON() ([]byte, error) {
	type alias SessionConfigOption
	type wire struct {
		*alias
		Options *SessionConfigSelectOptions `json:"options,omitempty"`
	}
	w := wire{alias: (*alias)(&o)}
	if !reflect.ValueOf(o.Options).IsZero() {
		Options := o.Options
		w.Options = &Options
	}
	switch o.Type {
	case SessionConfigOptionTypeSelect:
		Options := o.Options
		if Options.Ungrouped == nil && Options.Groups == nil {
			flat := UngroupedSessionConfigSelectOptions{}
			Options = SessionConfigSelectOptions{Ungrouped: &flat}
		}
		w.Options = &Options
	case SessionConfigOptionTypeBoolean:
	}
	return json.Marshal(w)
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *SessionConfigOption) UnmarshalJSON(data []byte) error {
	type alias SessionConfigOption
	decoded := alias{}
	raw := struct {
		Category     json.RawMessage `json:"category"`
		CurrentValue json.RawMessage `json:"currentValue"`
		Description  json.RawMessage `json:"description"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Category) > 0 {
		unmarshalDefault(raw.Category, &decoded.Category)
	}
	if len(raw.Description) > 0 {
		unmarshalDefault(raw.Description, &decoded.Description)
	}
	switch decoded.Type {
	case SessionConfigOptionTypeSelect:
		if err := requireJSONFieldsOnly(data, "type", "id", "name", "currentValue", "options"); err != nil {
			return err
		}
		var value SessionConfigValueID
		if err := json.Unmarshal(raw.CurrentValue, &value); err != nil {
			return err
		}
		decoded.CurrentValue = value
	case SessionConfigOptionTypeBoolean:
		if err := requireJSONFieldsOnly(data, "type", "id", "name", "currentValue"); err != nil {
			return err
		}
		var value bool
		if err := json.Unmarshal(raw.CurrentValue, &value); err != nil {
			return err
		}
		decoded.CurrentValue = value
	default:
		if err := requireJSONFieldsOnly(data, "type", "id", "name", "currentValue"); err != nil {
			return err
		}
		return invalidDiscriminator("type", decoded.Type)
	}
	*o = SessionConfigOption(decoded)
	return nil
}

// SessionConfigOptionCategory: Semantic category for a session configuration option.
//
// This is intended to help Clients distinguish broadly common selectors (e.g. model selector vs
// session mode selector vs thought/reasoning level) for UX purposes (keyboard shortcuts, icons,
// placement). It MUST NOT be required for correctness. Clients MUST handle missing or unknown
// categories gracefully.
//
// Category names beginning with `_` are free for custom use, like other ACP extension methods.
// Category names that do not begin with `_` are reserved for the ACP spec.
type SessionConfigOptionCategory string

const (
	// SessionConfigOptionCategoryMode: Session mode selector.
	SessionConfigOptionCategoryMode SessionConfigOptionCategory = "mode"
	// SessionConfigOptionCategoryModel: Model selector.
	SessionConfigOptionCategoryModel SessionConfigOptionCategory = "model"
	// SessionConfigOptionCategoryModelConfig: Model-related configuration parameter.
	SessionConfigOptionCategoryModelConfig SessionConfigOptionCategory = "model_config"
	// SessionConfigOptionCategoryThoughtLevel: Thought/reasoning level selector.
	SessionConfigOptionCategoryThoughtLevel SessionConfigOptionCategory = "thought_level"
)

// SessionConfigOptionsCapabilities: Session configuration option capabilities supported by the client.
type SessionConfigOptionsCapabilities struct {
	Meta    Meta                             `json:"_meta,omitzero"`
	Boolean *BooleanConfigOptionCapabilities `json:"boolean,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *SessionConfigOptionsCapabilities) UnmarshalJSON(data []byte) error {
	type alias SessionConfigOptionsCapabilities
	decoded := alias{}
	raw := struct {
		Boolean json.RawMessage `json:"boolean"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Boolean) > 0 {
		unmarshalDefault(raw.Boolean, &decoded.Boolean)
	}
	*c = SessionConfigOptionsCapabilities(decoded)
	return nil
}

// SessionConfigSelect: A single-value selector (dropdown) session configuration option payload.
type SessionConfigSelect struct {
	CurrentValue SessionConfigValueID       `json:"currentValue"`
	Options      SessionConfigSelectOptions `json:"options"`
}

// MarshalJSON implements json.Marshaler.
func (s SessionConfigSelect) MarshalJSON() ([]byte, error) {
	type alias SessionConfigSelect
	a := alias(s)
	if a.Options.Ungrouped == nil && a.Options.Groups == nil {
		flat := UngroupedSessionConfigSelectOptions{}
		a.Options = SessionConfigSelectOptions{Ungrouped: &flat}
	}
	return json.Marshal(a)
}

// SessionConfigSelectGroup: A group of possible values for a session configuration option.
type SessionConfigSelectGroup struct {
	Meta    Meta                        `json:"_meta,omitzero"`
	Group   SessionConfigGroupID        `json:"group"`
	Name    string                      `json:"name"`
	Options []SessionConfigSelectOption `json:"options"`
}

// MarshalJSON implements json.Marshaler.
func (g SessionConfigSelectGroup) MarshalJSON() ([]byte, error) {
	type alias SessionConfigSelectGroup
	a := alias(g)
	if a.Options == nil {
		a.Options = []SessionConfigSelectOption{}
	}
	return json.Marshal(a)
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *SessionConfigSelectGroup) UnmarshalJSON(data []byte) error {
	type alias SessionConfigSelectGroup
	decoded := alias{}
	raw := struct {
		Options json.RawMessage `json:"options"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Options) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.Options, &values); err == nil {
			decoded.Options = []SessionConfigSelectOption{}
			for _, value := range values {
				var item SessionConfigSelectOption
				if err := json.Unmarshal(value, &item); err == nil {
					decoded.Options = append(decoded.Options, item)
				}
			}
		}
	}
	if err := requireJSONFieldsOnly(data, "group", "name", "options"); err != nil {
		return err
	}
	*g = SessionConfigSelectGroup(decoded)
	return nil
}

// SessionConfigSelectOption: A possible value for a session configuration option.
type SessionConfigSelectOption struct {
	Meta        Meta                 `json:"_meta,omitzero"`
	Description *string              `json:"description,omitempty"`
	Name        string               `json:"name"`
	Value       SessionConfigValueID `json:"value"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *SessionConfigSelectOption) UnmarshalJSON(data []byte) error {
	type alias SessionConfigSelectOption
	decoded := alias{}
	raw := struct {
		Description json.RawMessage `json:"description"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Description) > 0 {
		unmarshalDefault(raw.Description, &decoded.Description)
	}
	if err := requireJSONFieldsOnly(data, "value", "name"); err != nil {
		return err
	}
	*o = SessionConfigSelectOption(decoded)
	return nil
}

// SessionConfigSelectOptions: Possible values for a session configuration option.
type SessionConfigSelectOptions struct {
	Ungrouped *UngroupedSessionConfigSelectOptions
	Groups    *GroupedSessionConfigSelectOptions
}

// UngroupedSessionConfigSelectOptions is the ungrouped variant of SessionConfigSelectOptions.
type UngroupedSessionConfigSelectOptions []SessionConfigSelectOption

// GroupedSessionConfigSelectOptions is the grouped variant of SessionConfigSelectOptions.
type GroupedSessionConfigSelectOptions []SessionConfigSelectGroup

// MarshalJSON implements json.Marshaler.
func (o SessionConfigSelectOptions) MarshalJSON() ([]byte, error) {
	if o.Groups != nil {
		return json.Marshal(o.Groups)
	}
	if o.Ungrouped != nil {
		return json.Marshal(o.Ungrouped)
	}
	return []byte("null"), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *SessionConfigSelectOptions) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*o = SessionConfigSelectOptions{}
		return nil
	}
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw) == 0 {
		flat := UngroupedSessionConfigSelectOptions{}
		*o = SessionConfigSelectOptions{Ungrouped: &flat}
		return nil
	}
	var probe struct {
		Group json.RawMessage `json:"group"`
	}
	if err := json.Unmarshal(raw[0], &probe); err != nil {
		return err
	}
	if len(probe.Group) > 0 {
		var groups GroupedSessionConfigSelectOptions
		if err := json.Unmarshal(data, &groups); err != nil {
			return err
		}
		*o = SessionConfigSelectOptions{Groups: &groups}
		return nil
	}
	var flat UngroupedSessionConfigSelectOptions
	if err := json.Unmarshal(data, &flat); err != nil {
		return err
	}
	*o = SessionConfigSelectOptions{Ungrouped: &flat}
	return nil
}

// SessionConfigValueID: Unique identifier for a session configuration option value.
type SessionConfigValueID string

// SessionDeleteCapabilities: Capabilities for the `session/delete` method.
//
// Supplying `{}` means the agent supports deleting sessions from `session/list`.
type SessionDeleteCapabilities struct {
	Meta Meta `json:"_meta,omitzero"`
}

// SessionID: A unique identifier for a conversation session between a client and agent.
//
// Sessions maintain their own context, conversation history, and state,
// allowing multiple independent interactions with the same agent.
//
// See protocol docs: [Session ID](https://agentclientprotocol.com/protocol/session-setup#session-id)
type SessionID string

// SessionInfo: Information about a session returned by session/list
type SessionInfo struct {
	Meta                  Meta      `json:"_meta,omitzero"`
	AdditionalDirectories []string  `json:"additionalDirectories,omitempty"`
	Cwd                   string    `json:"cwd"`
	SessionID             SessionID `json:"sessionId"`
	Title                 *string   `json:"title,omitempty"`
	UpdatedAt             *string   `json:"updatedAt,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (i *SessionInfo) UnmarshalJSON(data []byte) error {
	type alias SessionInfo
	decoded := alias{}
	raw := struct {
		AdditionalDirectories json.RawMessage `json:"additionalDirectories"`
		Title                 json.RawMessage `json:"title"`
		UpdatedAt             json.RawMessage `json:"updatedAt"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.AdditionalDirectories) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.AdditionalDirectories, &values); err == nil {
			decoded.AdditionalDirectories = []string{}
			for _, value := range values {
				var item string
				if err := json.Unmarshal(value, &item); err == nil {
					decoded.AdditionalDirectories = append(decoded.AdditionalDirectories, item)
				}
			}
		}
	}
	if len(raw.Title) > 0 {
		unmarshalDefault(raw.Title, &decoded.Title)
	}
	if len(raw.UpdatedAt) > 0 {
		unmarshalDefault(raw.UpdatedAt, &decoded.UpdatedAt)
	}
	if err := requireJSONFieldsOnly(data, "sessionId", "cwd"); err != nil {
		return err
	}
	*i = SessionInfo(decoded)
	return nil
}

// SessionInfoUpdate: Update to session metadata. All fields are optional to support partial updates.
//
// Agents send this notification to update session information like title or custom metadata.
// This allows clients to display dynamic session names and track session state changes.
type SessionInfoUpdate struct {
	Meta      Meta    `json:"_meta,omitzero"`
	Title     *string `json:"title,omitempty"`
	UpdatedAt *string `json:"updatedAt,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (u *SessionInfoUpdate) UnmarshalJSON(data []byte) error {
	type alias SessionInfoUpdate
	decoded := alias{}
	raw := struct {
		Title     json.RawMessage `json:"title"`
		UpdatedAt json.RawMessage `json:"updatedAt"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Title) > 0 {
		unmarshalDefault(raw.Title, &decoded.Title)
	}
	if len(raw.UpdatedAt) > 0 {
		unmarshalDefault(raw.UpdatedAt, &decoded.UpdatedAt)
	}
	*u = SessionInfoUpdate(decoded)
	return nil
}

// SessionListCapabilities: Capabilities for the `session/list` method.
//
// Supplying `{}` means the agent supports listing sessions.
type SessionListCapabilities struct {
	Meta Meta `json:"_meta,omitzero"`
}

// SessionMode: A mode the agent can operate in.
//
// See protocol docs: [Session Modes](https://agentclientprotocol.com/protocol/session-modes)
type SessionMode struct {
	Meta        Meta          `json:"_meta,omitzero"`
	Description *string       `json:"description,omitempty"`
	ID          SessionModeID `json:"id"`
	Name        string        `json:"name"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *SessionMode) UnmarshalJSON(data []byte) error {
	type alias SessionMode
	decoded := alias{}
	raw := struct {
		Description json.RawMessage `json:"description"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Description) > 0 {
		unmarshalDefault(raw.Description, &decoded.Description)
	}
	if err := requireJSONFieldsOnly(data, "id", "name"); err != nil {
		return err
	}
	*m = SessionMode(decoded)
	return nil
}

// SessionModeID: Unique identifier for a Session Mode.
type SessionModeID string

// SessionModeState: The set of modes and the one currently active.
type SessionModeState struct {
	Meta           Meta          `json:"_meta,omitzero"`
	AvailableModes []SessionMode `json:"availableModes"`
	CurrentModeID  SessionModeID `json:"currentModeId"`
}

// MarshalJSON implements json.Marshaler.
func (s SessionModeState) MarshalJSON() ([]byte, error) {
	type alias SessionModeState
	a := alias(s)
	if a.AvailableModes == nil {
		a.AvailableModes = []SessionMode{}
	}
	return json.Marshal(a)
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *SessionModeState) UnmarshalJSON(data []byte) error {
	type alias SessionModeState
	decoded := alias{}
	raw := struct {
		AvailableModes json.RawMessage `json:"availableModes"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.AvailableModes) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.AvailableModes, &values); err == nil {
			decoded.AvailableModes = []SessionMode{}
			for _, value := range values {
				var item SessionMode
				if err := json.Unmarshal(value, &item); err == nil {
					decoded.AvailableModes = append(decoded.AvailableModes, item)
				}
			}
		}
	}
	if err := requireJSONFieldsOnly(data, "currentModeId"); err != nil {
		return err
	}
	if err := requireJSONKeysOnly(data, "availableModes"); err != nil {
		return err
	}
	*s = SessionModeState(decoded)
	return nil
}

// SessionNotification: Notification containing a session update from the agent.
//
// Used to stream real-time progress and results during prompt processing.
//
// See protocol docs: [Agent Reports Output](https://agentclientprotocol.com/protocol/prompt-turn#3-agent-reports-output)
type SessionNotification struct {
	Meta      Meta          `json:"_meta,omitzero"`
	SessionID SessionID     `json:"sessionId"`
	Update    SessionUpdate `json:"update"`
}

// SessionResumeCapabilities: Capabilities for the `session/resume` method.
//
// Supplying `{}` means the agent supports resuming sessions.
type SessionResumeCapabilities struct {
	Meta Meta `json:"_meta,omitzero"`
}

// SessionUpdate: Different types of updates that can be sent during session processing.
//
// These updates provide real-time feedback about the agent's progress.
//
// See protocol docs: [Agent Reports Output](https://agentclientprotocol.com/protocol/prompt-turn#3-agent-reports-output)
type SessionUpdate struct {
	SessionUpdate     SessionUpdateType     `json:"sessionUpdate"`
	Meta              Meta                  `json:"_meta,omitzero"`
	AvailableCommands []AvailableCommand    `json:"availableCommands,omitempty"`
	ConfigOptions     []SessionConfigOption `json:"configOptions,omitempty"`
	Content           any                   `json:"content,omitempty,omitzero"`
	Cost              *Cost                 `json:"cost,omitempty"`
	CurrentModeID     SessionModeID         `json:"currentModeId,omitempty"`
	Entries           []PlanEntry           `json:"entries,omitempty"`
	Kind              *ToolKind             `json:"kind,omitempty"`
	Locations         *[]ToolCallLocation   `json:"locations,omitempty"`
	MessageID         *MessageID            `json:"messageId,omitempty"`
	RawInput          any                   `json:"rawInput,omitempty"`
	RawOutput         any                   `json:"rawOutput,omitempty"`
	Size              uint64                `json:"size,omitempty"`
	Status            *ToolCallStatus       `json:"status,omitempty"`
	Title             *string               `json:"title,omitempty"`
	ToolCallID        ToolCallID            `json:"toolCallId,omitempty"`
	UpdatedAt         *string               `json:"updatedAt,omitempty"`
	Used              uint64                `json:"used,omitempty"`
}

// SessionUpdateType is the discriminator for SessionUpdate variants.
type SessionUpdateType string

const (
	SessionUpdateTypeUserMessageChunk        SessionUpdateType = "user_message_chunk"
	SessionUpdateTypeAgentMessageChunk       SessionUpdateType = "agent_message_chunk"
	SessionUpdateTypeAgentThoughtChunk       SessionUpdateType = "agent_thought_chunk"
	SessionUpdateTypeToolCall                SessionUpdateType = "tool_call"
	SessionUpdateTypeToolCallUpdate          SessionUpdateType = "tool_call_update"
	SessionUpdateTypePlan                    SessionUpdateType = "plan"
	SessionUpdateTypeAvailableCommandsUpdate SessionUpdateType = "available_commands_update"
	SessionUpdateTypeCurrentModeUpdate       SessionUpdateType = "current_mode_update"
	SessionUpdateTypeConfigOptionUpdate      SessionUpdateType = "config_option_update"
	SessionUpdateTypeSessionInfoUpdate       SessionUpdateType = "session_info_update"
	SessionUpdateTypeUsageUpdate             SessionUpdateType = "usage_update"
)

// UserMessageChunkSessionUpdate creates an SessionUpdate variant: A chunk of the user's message being streamed.
func UserMessageChunkSessionUpdate(content ContentBlock) SessionUpdate {
	return SessionUpdate{
		SessionUpdate: SessionUpdateTypeUserMessageChunk,
		Content:       content,
	}
}

// AgentMessageChunkSessionUpdate creates an SessionUpdate variant: A chunk of the agent's response being streamed.
func AgentMessageChunkSessionUpdate(content ContentBlock) SessionUpdate {
	return SessionUpdate{
		SessionUpdate: SessionUpdateTypeAgentMessageChunk,
		Content:       content,
	}
}

// AgentThoughtChunkSessionUpdate creates an SessionUpdate variant: A chunk of the agent's internal reasoning being streamed.
func AgentThoughtChunkSessionUpdate(content ContentBlock) SessionUpdate {
	return SessionUpdate{
		SessionUpdate: SessionUpdateTypeAgentThoughtChunk,
		Content:       content,
	}
}

// ToolCallSessionUpdate creates an SessionUpdate variant: Notification that a new tool call has been initiated.
func ToolCallSessionUpdate(toolCallID ToolCallID, title string) SessionUpdate {
	return SessionUpdate{
		SessionUpdate: SessionUpdateTypeToolCall,
		Title:         &title,
		ToolCallID:    toolCallID,
	}
}

// ToolCallUpdateSessionUpdate creates an SessionUpdate variant: Update on the status or results of a tool call.
func ToolCallUpdateSessionUpdate(toolCallID ToolCallID) SessionUpdate {
	return SessionUpdate{
		SessionUpdate: SessionUpdateTypeToolCallUpdate,
		ToolCallID:    toolCallID,
	}
}

// PlanSessionUpdate creates an SessionUpdate variant: The agent's execution plan for complex tasks.
// See protocol docs: [Agent Plan](https://agentclientprotocol.com/protocol/agent-plan)
func PlanSessionUpdate(entries []PlanEntry) SessionUpdate {
	return SessionUpdate{
		SessionUpdate: SessionUpdateTypePlan,
		Entries:       entries,
	}
}

// AvailableCommandsUpdateSessionUpdate creates an SessionUpdate variant: Available commands are ready or have changed
func AvailableCommandsUpdateSessionUpdate(availableCommands []AvailableCommand) SessionUpdate {
	return SessionUpdate{
		SessionUpdate:     SessionUpdateTypeAvailableCommandsUpdate,
		AvailableCommands: availableCommands,
	}
}

// CurrentModeUpdateSessionUpdate creates an SessionUpdate variant: The current mode of the session has changed
//
// See protocol docs: [Session Modes](https://agentclientprotocol.com/protocol/session-modes)
func CurrentModeUpdateSessionUpdate(currentModeID SessionModeID) SessionUpdate {
	return SessionUpdate{
		SessionUpdate: SessionUpdateTypeCurrentModeUpdate,
		CurrentModeID: currentModeID,
	}
}

// ConfigOptionUpdateSessionUpdate creates an SessionUpdate variant: Session configuration options have been updated.
func ConfigOptionUpdateSessionUpdate(configOptions []SessionConfigOption) SessionUpdate {
	return SessionUpdate{
		SessionUpdate: SessionUpdateTypeConfigOptionUpdate,
		ConfigOptions: configOptions,
	}
}

// SessionInfoSessionUpdate creates an SessionUpdate variant: Session metadata has been updated (title, timestamps, custom metadata)
func SessionInfoSessionUpdate() SessionUpdate {
	return SessionUpdate{
		SessionUpdate: SessionUpdateTypeSessionInfoUpdate,
	}
}

// UsageUpdateSessionUpdate creates an SessionUpdate variant: Context window and cost update for the session.
func UsageUpdateSessionUpdate(used uint64, size uint64) SessionUpdate {
	return SessionUpdate{
		SessionUpdate: SessionUpdateTypeUsageUpdate,
		Size:          size,
		Used:          used,
	}
}

// MarshalJSON implements json.Marshaler.
func (u SessionUpdate) MarshalJSON() ([]byte, error) {
	type alias SessionUpdate
	// Use the pointer-heavy wire form only when a required zero value must
	// override omitempty.
	switch u.SessionUpdate {
	case SessionUpdateTypeUserMessageChunk, SessionUpdateTypeAgentMessageChunk, SessionUpdateTypeAgentThoughtChunk:
		if u.Content != nil {
			return json.Marshal((*alias)(&u))
		}
	case SessionUpdateTypeToolCall:
		if u.Title != nil && u.ToolCallID != "" {
			return json.Marshal((*alias)(&u))
		}
	case SessionUpdateTypeToolCallUpdate:
		if u.ToolCallID != "" {
			return json.Marshal((*alias)(&u))
		}
	case SessionUpdateTypePlan:
		if len(u.Entries) > 0 {
			return json.Marshal((*alias)(&u))
		}
	case SessionUpdateTypeAvailableCommandsUpdate:
		if len(u.AvailableCommands) > 0 {
			return json.Marshal((*alias)(&u))
		}
	case SessionUpdateTypeCurrentModeUpdate:
		if u.CurrentModeID != "" {
			return json.Marshal((*alias)(&u))
		}
	case SessionUpdateTypeConfigOptionUpdate:
		if len(u.ConfigOptions) > 0 {
			return json.Marshal((*alias)(&u))
		}
	case SessionUpdateTypeSessionInfoUpdate:
		return json.Marshal((*alias)(&u))
	case SessionUpdateTypeUsageUpdate:
		if u.Size != 0 && u.Used != 0 {
			return json.Marshal((*alias)(&u))
		}
	}
	type wire struct {
		*alias
		Content           *any                   `json:"content,omitempty"`
		Title             **string               `json:"title,omitempty"`
		ToolCallID        *ToolCallID            `json:"toolCallId,omitempty"`
		Entries           *[]PlanEntry           `json:"entries,omitempty"`
		AvailableCommands *[]AvailableCommand    `json:"availableCommands,omitempty"`
		CurrentModeID     *SessionModeID         `json:"currentModeId,omitempty"`
		ConfigOptions     *[]SessionConfigOption `json:"configOptions,omitempty"`
		Size              *uint64                `json:"size,omitempty"`
		Used              *uint64                `json:"used,omitempty"`
	}
	w := wire{alias: (*alias)(&u)}
	if u.Content != nil {
		Content := u.Content
		w.Content = &Content
	}
	if u.Title != nil {
		Title := u.Title
		w.Title = &Title
	}
	if !reflect.ValueOf(u.ToolCallID).IsZero() {
		toolCallID := u.ToolCallID
		w.ToolCallID = &toolCallID
	}
	if len(u.Entries) > 0 {
		Entries := u.Entries
		w.Entries = &Entries
	}
	if len(u.AvailableCommands) > 0 {
		AvailableCommands := u.AvailableCommands
		w.AvailableCommands = &AvailableCommands
	}
	if !reflect.ValueOf(u.CurrentModeID).IsZero() {
		CurrentModeID := u.CurrentModeID
		w.CurrentModeID = &CurrentModeID
	}
	if len(u.ConfigOptions) > 0 {
		ConfigOptions := u.ConfigOptions
		w.ConfigOptions = &ConfigOptions
	}
	if !reflect.ValueOf(u.Size).IsZero() {
		Size := u.Size
		w.Size = &Size
	}
	if !reflect.ValueOf(u.Used).IsZero() {
		Used := u.Used
		w.Used = &Used
	}
	switch u.SessionUpdate {
	case SessionUpdateTypeUserMessageChunk:
		Content := u.Content
		w.Content = &Content
	case SessionUpdateTypeAgentMessageChunk:
		Content := u.Content
		w.Content = &Content
	case SessionUpdateTypeAgentThoughtChunk:
		Content := u.Content
		w.Content = &Content
	case SessionUpdateTypeToolCall:
		Title := u.Title
		w.Title = &Title
		toolCallID := u.ToolCallID
		w.ToolCallID = &toolCallID
	case SessionUpdateTypeToolCallUpdate:
		toolCallID := u.ToolCallID
		w.ToolCallID = &toolCallID
	case SessionUpdateTypePlan:
		Entries := u.Entries
		if Entries == nil {
			Entries = []PlanEntry{}
		}
		w.Entries = &Entries
	case SessionUpdateTypeAvailableCommandsUpdate:
		AvailableCommands := u.AvailableCommands
		if AvailableCommands == nil {
			AvailableCommands = []AvailableCommand{}
		}
		w.AvailableCommands = &AvailableCommands
	case SessionUpdateTypeCurrentModeUpdate:
		CurrentModeID := u.CurrentModeID
		w.CurrentModeID = &CurrentModeID
	case SessionUpdateTypeConfigOptionUpdate:
		ConfigOptions := u.ConfigOptions
		if ConfigOptions == nil {
			ConfigOptions = []SessionConfigOption{}
		}
		w.ConfigOptions = &ConfigOptions
	case SessionUpdateTypeUsageUpdate:
		Size := u.Size
		w.Size = &Size
		Used := u.Used
		w.Used = &Used
	case SessionUpdateTypeSessionInfoUpdate:
	}
	return json.Marshal(w)
}

// UnmarshalJSON implements json.Unmarshaler.
func (u *SessionUpdate) UnmarshalJSON(data []byte) error {
	type alias SessionUpdate
	decoded := alias{}
	raw := struct {
		AvailableCommands json.RawMessage `json:"availableCommands"`
		ConfigOptions     json.RawMessage `json:"configOptions"`
		Content           json.RawMessage `json:"content"`
		Cost              json.RawMessage `json:"cost"`
		Entries           json.RawMessage `json:"entries"`
		Kind              json.RawMessage `json:"kind"`
		Locations         json.RawMessage `json:"locations"`
		MessageID         json.RawMessage `json:"messageId"`
		RawInput          json.RawMessage `json:"rawInput"`
		RawOutput         json.RawMessage `json:"rawOutput"`
		Status            json.RawMessage `json:"status"`
		Title             json.RawMessage `json:"title"`
		UpdatedAt         json.RawMessage `json:"updatedAt"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.AvailableCommands) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.AvailableCommands, &values); err == nil {
			decoded.AvailableCommands = []AvailableCommand{}
			for _, value := range values {
				var item AvailableCommand
				if err := json.Unmarshal(value, &item); err == nil {
					decoded.AvailableCommands = append(decoded.AvailableCommands, item)
				}
			}
		}
	}
	if len(raw.ConfigOptions) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.ConfigOptions, &values); err == nil {
			decoded.ConfigOptions = []SessionConfigOption{}
			for _, value := range values {
				var item SessionConfigOption
				if err := json.Unmarshal(value, &item); err == nil {
					switch item.Type {
					case SessionConfigOptionTypeSelect, SessionConfigOptionTypeBoolean:
						decoded.ConfigOptions = append(decoded.ConfigOptions, item)
					}
				}
			}
		}
	}
	if len(raw.Content) > 0 {
		switch decoded.SessionUpdate {
		case SessionUpdateTypeUserMessageChunk, SessionUpdateTypeAgentMessageChunk, SessionUpdateTypeAgentThoughtChunk:
			var content ContentBlock
			if err := json.Unmarshal(raw.Content, &content); err != nil {
				return err
			}
			decoded.Content = content
		case SessionUpdateTypeToolCall, SessionUpdateTypeToolCallUpdate:
			isToolCallUpdate := decoded.SessionUpdate == SessionUpdateTypeToolCallUpdate
			if isToolCallUpdate && bytes.Equal(bytes.TrimSpace(raw.Content), []byte("null")) {
				decoded.Content = json.RawMessage("null")
				break
			}
			var values []json.RawMessage
			if err := json.Unmarshal(raw.Content, &values); err == nil && values != nil {
				items := []ToolCallContent{}
				for _, value := range values {
					var item ToolCallContent
					if err := json.Unmarshal(value, &item); err == nil {
						items = append(items, item)
					}
				}
				decoded.Content = items
			}
		case SessionUpdateTypePlan,
			SessionUpdateTypeAvailableCommandsUpdate,
			SessionUpdateTypeCurrentModeUpdate,
			SessionUpdateTypeConfigOptionUpdate,
			SessionUpdateTypeSessionInfoUpdate,
			SessionUpdateTypeUsageUpdate:
		}
	}
	if len(raw.Cost) > 0 {
		unmarshalDefault(raw.Cost, &decoded.Cost)
	}
	if len(raw.Entries) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.Entries, &values); err == nil {
			decoded.Entries = []PlanEntry{}
			for _, value := range values {
				var item PlanEntry
				if err := json.Unmarshal(value, &item); err == nil {
					decoded.Entries = append(decoded.Entries, item)
				}
			}
		}
	}
	if len(raw.Kind) > 0 {
		var value ToolKind
		if err := json.Unmarshal(raw.Kind, &value); err == nil {
			switch string(value) {
			case "read", "edit", "delete", "move", "search", "execute", "think", "fetch", "switch_mode", "other":
				decoded.Kind = &value
			}
		}
	}
	if len(raw.Locations) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.Locations, &values); err == nil && values != nil {
			items := []ToolCallLocation{}
			for _, value := range values {
				var item ToolCallLocation
				if err := json.Unmarshal(value, &item); err == nil {
					items = append(items, item)
				}
			}
			decoded.Locations = &items
		}
	}
	if len(raw.MessageID) > 0 {
		unmarshalDefault(raw.MessageID, &decoded.MessageID)
	}
	if len(raw.RawInput) > 0 {
		unmarshalDefault(raw.RawInput, &decoded.RawInput)
	}
	if len(raw.RawOutput) > 0 {
		unmarshalDefault(raw.RawOutput, &decoded.RawOutput)
	}
	if len(raw.Status) > 0 {
		var value ToolCallStatus
		if err := json.Unmarshal(raw.Status, &value); err == nil {
			switch string(value) {
			case "pending", "in_progress", "completed", "failed":
				decoded.Status = &value
			}
		}
	}
	if len(raw.Title) > 0 {
		unmarshalDefault(raw.Title, &decoded.Title)
	}
	if len(raw.UpdatedAt) > 0 {
		unmarshalDefault(raw.UpdatedAt, &decoded.UpdatedAt)
	}
	switch decoded.SessionUpdate {
	case SessionUpdateTypeUserMessageChunk, SessionUpdateTypeAgentMessageChunk, SessionUpdateTypeAgentThoughtChunk:
		if err := requireJSONFieldsOnly(data, "content"); err != nil {
			return err
		}
	case SessionUpdateTypeToolCall:
		if err := requireJSONFieldsOnly(data, "toolCallId", "title"); err != nil {
			return err
		}
		var title string
		if err := json.Unmarshal(raw.Title, &title); err != nil {
			return err
		}
		decoded.Title = &title
	case SessionUpdateTypeToolCallUpdate:
		if err := requireJSONFieldsOnly(data, "toolCallId"); err != nil {
			return err
		}
	case SessionUpdateTypePlan:
		if err := requireJSONKeysOnly(data, "entries"); err != nil {
			return err
		}
	case SessionUpdateTypeAvailableCommandsUpdate:
		if err := requireJSONKeysOnly(data, "availableCommands"); err != nil {
			return err
		}
	case SessionUpdateTypeCurrentModeUpdate:
		if err := requireJSONFieldsOnly(data, "currentModeId"); err != nil {
			return err
		}
	case SessionUpdateTypeConfigOptionUpdate:
		if err := requireJSONKeysOnly(data, "configOptions"); err != nil {
			return err
		}
	case SessionUpdateTypeSessionInfoUpdate:
	case SessionUpdateTypeUsageUpdate:
		if err := requireJSONFieldsOnly(data, "used", "size"); err != nil {
			return err
		}
	default:
		return invalidDiscriminator("sessionUpdate", decoded.SessionUpdate)
	}
	*u = SessionUpdate(decoded)
	return nil
}

// SetSessionConfigOptionRequest: Request parameters for setting a session configuration option.
type SetSessionConfigOptionRequest struct {
	Type      SetSessionConfigOptionRequestType `json:"type,omitempty"`
	Meta      Meta                              `json:"_meta,omitzero"`
	ConfigID  SessionConfigID                   `json:"configId"`
	SessionID SessionID                         `json:"sessionId"`
	Value     any                               `json:"value"`
}

// SetSessionConfigOptionRequestType is the discriminator for SetSessionConfigOptionRequest variants.
type SetSessionConfigOptionRequestType string

const (
	SetSessionConfigOptionRequestTypeBoolean SetSessionConfigOptionRequestType = "boolean"
)

// BooleanSetSessionConfigOptionRequest creates a boolean value (`type: "boolean"`).
func BooleanSetSessionConfigOptionRequest(sessionID SessionID, configID SessionConfigID, value bool) SetSessionConfigOptionRequest {
	return SetSessionConfigOptionRequest{
		Type:      SetSessionConfigOptionRequestTypeBoolean,
		ConfigID:  configID,
		SessionID: sessionID,
		Value:     value,
	}
}

// ValueIDSetSessionConfigOptionRequest creates a [`SessionConfigValueID`] string value.
//
// This is the default when `type` is absent on the wire. Unknown `type`
// values with string payloads also gracefully deserialize into this
// variant.
func ValueIDSetSessionConfigOptionRequest(sessionID SessionID, configID SessionConfigID, value SessionConfigValueID) SetSessionConfigOptionRequest {
	return SetSessionConfigOptionRequest{
		ConfigID:  configID,
		SessionID: sessionID,
		Value:     value,
	}
}

// SetSessionConfigOptionResponse: Response to `session/set_config_option` method.
type SetSessionConfigOptionResponse struct {
	Meta          Meta                  `json:"_meta,omitzero"`
	ConfigOptions []SessionConfigOption `json:"configOptions"`
}

// MarshalJSON implements json.Marshaler.
func (r SetSessionConfigOptionResponse) MarshalJSON() ([]byte, error) {
	type alias SetSessionConfigOptionResponse
	a := alias(r)
	if a.ConfigOptions == nil {
		a.ConfigOptions = []SessionConfigOption{}
	}
	return json.Marshal(a)
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *SetSessionConfigOptionResponse) UnmarshalJSON(data []byte) error {
	type alias SetSessionConfigOptionResponse
	decoded := alias{}
	raw := struct {
		ConfigOptions json.RawMessage `json:"configOptions"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.ConfigOptions) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.ConfigOptions, &values); err == nil {
			decoded.ConfigOptions = []SessionConfigOption{}
			for _, value := range values {
				var item SessionConfigOption
				if err := json.Unmarshal(value, &item); err == nil {
					switch item.Type {
					case SessionConfigOptionTypeSelect, SessionConfigOptionTypeBoolean:
						decoded.ConfigOptions = append(decoded.ConfigOptions, item)
					}
				}
			}
		}
	}
	if err := requireJSONKeysOnly(data, "configOptions"); err != nil {
		return err
	}
	*r = SetSessionConfigOptionResponse(decoded)
	return nil
}

// SetSessionModeRequest: Request parameters for setting a session mode.
type SetSessionModeRequest struct {
	Meta      Meta          `json:"_meta,omitzero"`
	ModeID    SessionModeID `json:"modeId"`
	SessionID SessionID     `json:"sessionId"`
}

// SetSessionModeResponse: Response to `session/set_mode` method.
type SetSessionModeResponse struct {
	Meta Meta `json:"_meta,omitzero"`
}

// StopReason: Reasons why an agent stops processing a prompt turn.
//
// See protocol docs: [Stop Reasons](https://agentclientprotocol.com/protocol/prompt-turn#stop-reasons)
type StopReason string

const (
	// StopReasonEndTurn: The turn ended successfully.
	StopReasonEndTurn StopReason = "end_turn"
	// StopReasonMaxTokens: The turn ended because the agent reached the maximum number of tokens.
	StopReasonMaxTokens StopReason = "max_tokens"
	// StopReasonMaxTurnRequests: The turn ended because the agent reached the maximum number of allowed
	// agent requests between user turns.
	StopReasonMaxTurnRequests StopReason = "max_turn_requests"
	// StopReasonRefusal: The turn ended because the agent refused to continue. The user prompt
	// and everything that comes after it won't be included in the next
	// prompt, so this should be reflected in the UI.
	StopReasonRefusal StopReason = "refusal"
	// StopReasonCanceled: The turn was cancelled by the client via `session/cancel`.
	//
	// This stop reason MUST be returned when the client sends a `session/cancel`
	// notification, even if the cancellation causes exceptions in underlying operations.
	// Agents should catch these exceptions and return this semantically meaningful
	// response to confirm successful cancellation.
	StopReasonCanceled StopReason = "cancelled"
)

// StringFormat: String format types for string properties in elicitation schemas.
type StringFormat string

const (
	// StringFormatEmail: Email address format.
	StringFormatEmail StringFormat = "email"
	// StringFormatURI: URI format.
	StringFormatURI StringFormat = "uri"
	// StringFormatDate: Date format (YYYY-MM-DD).
	StringFormatDate StringFormat = "date"
	// StringFormatDateTime: Date-time format (ISO 8601).
	StringFormatDateTime StringFormat = "date-time"
)

// StringMultiSelectItems: String item schema for multi-select enum properties.
type StringMultiSelectItems struct {
	Meta Meta     `json:"_meta,omitzero"`
	Enum []string `json:"enum"`
}

// MarshalJSON implements json.Marshaler.
func (i StringMultiSelectItems) MarshalJSON() ([]byte, error) {
	type alias StringMultiSelectItems
	a := alias(i)
	if a.Enum == nil {
		a.Enum = []string{}
	}
	return json.Marshal(a)
}

// StringPropertySchema: Schema for string properties in an elicitation form.
//
// When `enum` or `oneOf` is set, this represents a single-select enum
// with `"type": "string"`.
type StringPropertySchema struct {
	Meta        Meta          `json:"_meta,omitzero"`
	Default     *string       `json:"default,omitempty"`
	Description *string       `json:"description,omitempty"`
	Enum        *[]string     `json:"enum,omitempty"`
	Format      *StringFormat `json:"format,omitempty"`
	MaxLength   *uint32       `json:"maxLength,omitempty"`
	MinLength   *uint32       `json:"minLength,omitempty"`
	OneOf       *[]EnumOption `json:"oneOf,omitempty"`
	Pattern     *string       `json:"pattern,omitempty"`
	Title       *string       `json:"title,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *StringPropertySchema) UnmarshalJSON(data []byte) error {
	type alias StringPropertySchema
	decoded := alias{}
	raw := struct {
		Default     json.RawMessage `json:"default"`
		Description json.RawMessage `json:"description"`
		Title       json.RawMessage `json:"title"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Default) > 0 {
		unmarshalDefault(raw.Default, &decoded.Default)
	}
	if len(raw.Description) > 0 {
		unmarshalDefault(raw.Description, &decoded.Description)
	}
	if len(raw.Title) > 0 {
		unmarshalDefault(raw.Title, &decoded.Title)
	}
	*s = StringPropertySchema(decoded)
	return nil
}

// Terminal: Embed a terminal created with `terminal/create` by its id.
//
// The terminal must be added before calling `terminal/release`.
//
// See protocol docs: [Terminal](https://agentclientprotocol.com/protocol/terminals)
type Terminal struct {
	Meta       Meta       `json:"_meta,omitzero"`
	TerminalID TerminalID `json:"terminalId"`
}

// TerminalExitStatus: Exit status of a terminal command.
type TerminalExitStatus struct {
	Meta     Meta    `json:"_meta,omitzero"`
	ExitCode *uint32 `json:"exitCode,omitempty"`
	Signal   *string `json:"signal,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *TerminalExitStatus) UnmarshalJSON(data []byte) error {
	type alias TerminalExitStatus
	decoded := alias{}
	raw := struct {
		ExitCode json.RawMessage `json:"exitCode"`
		Signal   json.RawMessage `json:"signal"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.ExitCode) > 0 {
		unmarshalDefault(raw.ExitCode, &decoded.ExitCode)
	}
	if len(raw.Signal) > 0 {
		unmarshalDefault(raw.Signal, &decoded.Signal)
	}
	*s = TerminalExitStatus(decoded)
	return nil
}

// TerminalID: Typed identifier used for terminal values on the wire.
type TerminalID string

// TerminalOutputRequest: Request to get the current output and status of a terminal.
type TerminalOutputRequest struct {
	Meta       Meta       `json:"_meta,omitzero"`
	SessionID  SessionID  `json:"sessionId"`
	TerminalID TerminalID `json:"terminalId"`
}

// TerminalOutputResponse: Response containing the terminal output and exit status.
type TerminalOutputResponse struct {
	Meta       Meta                `json:"_meta,omitzero"`
	ExitStatus *TerminalExitStatus `json:"exitStatus,omitempty"`
	Output     string              `json:"output"`
	Truncated  bool                `json:"truncated"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *TerminalOutputResponse) UnmarshalJSON(data []byte) error {
	type alias TerminalOutputResponse
	decoded := alias{}
	raw := struct {
		ExitStatus json.RawMessage `json:"exitStatus"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.ExitStatus) > 0 {
		unmarshalDefault(raw.ExitStatus, &decoded.ExitStatus)
	}
	*r = TerminalOutputResponse(decoded)
	return nil
}

// TextContent: Text provided to or from an LLM.
type TextContent struct {
	Meta        Meta         `json:"_meta,omitzero"`
	Annotations *Annotations `json:"annotations,omitempty"`
	Text        string       `json:"text"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *TextContent) UnmarshalJSON(data []byte) error {
	type alias TextContent
	decoded := alias{}
	raw := struct {
		Annotations json.RawMessage `json:"annotations"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Annotations) > 0 {
		unmarshalDefault(raw.Annotations, &decoded.Annotations)
	}
	if err := requireJSONFieldsOnly(data, "text"); err != nil {
		return err
	}
	*c = TextContent(decoded)
	return nil
}

// TextResourceContents: Text-based resource contents.
type TextResourceContents struct {
	Meta     Meta    `json:"_meta,omitzero"`
	MIMEType *string `json:"mimeType,omitempty"`
	Text     string  `json:"text"`
	URI      string  `json:"uri"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *TextResourceContents) UnmarshalJSON(data []byte) error {
	type alias TextResourceContents
	decoded := alias{}
	raw := struct {
		MIMEType json.RawMessage `json:"mimeType"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.MIMEType) > 0 {
		unmarshalDefault(raw.MIMEType, &decoded.MIMEType)
	}
	if err := requireJSONFieldsOnly(data, "text", "uri"); err != nil {
		return err
	}
	*c = TextResourceContents(decoded)
	return nil
}

// TitledMultiSelectItems: Items definition for titled multi-select enum properties.
type TitledMultiSelectItems struct {
	Meta  Meta         `json:"_meta,omitzero"`
	AnyOf []EnumOption `json:"anyOf"`
}

// MarshalJSON implements json.Marshaler.
func (i TitledMultiSelectItems) MarshalJSON() ([]byte, error) {
	type alias TitledMultiSelectItems
	a := alias(i)
	if a.AnyOf == nil {
		a.AnyOf = []EnumOption{}
	}
	return json.Marshal(a)
}

// ToolCall: Represents a tool call that the language model has requested.
//
// Tool calls are actions that the agent executes on behalf of the language model,
// such as reading files, executing code, or fetching data from external sources.
//
// See protocol docs: [Tool Calls](https://agentclientprotocol.com/protocol/tool-calls)
type ToolCall struct {
	Meta       Meta               `json:"_meta,omitzero"`
	Content    []ToolCallContent  `json:"content,omitempty"`
	Kind       ToolKind           `json:"kind,omitempty"`
	Locations  []ToolCallLocation `json:"locations,omitempty"`
	RawInput   any                `json:"rawInput,omitempty"`
	RawOutput  any                `json:"rawOutput,omitempty"`
	Status     ToolCallStatus     `json:"status,omitempty"`
	Title      string             `json:"title"`
	ToolCallID ToolCallID         `json:"toolCallId"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *ToolCall) UnmarshalJSON(data []byte) error {
	type alias ToolCall
	decoded := alias{}
	raw := struct {
		Content   json.RawMessage `json:"content"`
		Kind      json.RawMessage `json:"kind"`
		Locations json.RawMessage `json:"locations"`
		RawInput  json.RawMessage `json:"rawInput"`
		RawOutput json.RawMessage `json:"rawOutput"`
		Status    json.RawMessage `json:"status"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Content) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.Content, &values); err == nil {
			decoded.Content = []ToolCallContent{}
			for _, value := range values {
				var item ToolCallContent
				if err := json.Unmarshal(value, &item); err == nil {
					switch item.Type {
					case ToolCallContentTypeContent, ToolCallContentTypeDiff, ToolCallContentTypeTerminal:
						decoded.Content = append(decoded.Content, item)
					}
				}
			}
		}
	}
	if len(raw.Kind) > 0 {
		var value ToolKind
		if err := json.Unmarshal(raw.Kind, &value); err == nil {
			switch string(value) {
			case "read", "edit", "delete", "move", "search", "execute", "think", "fetch", "switch_mode", "other":
				decoded.Kind = value
			}
		}
	}
	if len(raw.Locations) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.Locations, &values); err == nil {
			decoded.Locations = []ToolCallLocation{}
			for _, value := range values {
				var item ToolCallLocation
				if err := json.Unmarshal(value, &item); err == nil {
					decoded.Locations = append(decoded.Locations, item)
				}
			}
		}
	}
	if len(raw.RawInput) > 0 {
		unmarshalDefault(raw.RawInput, &decoded.RawInput)
	}
	if len(raw.RawOutput) > 0 {
		unmarshalDefault(raw.RawOutput, &decoded.RawOutput)
	}
	if len(raw.Status) > 0 {
		var value ToolCallStatus
		if err := json.Unmarshal(raw.Status, &value); err == nil {
			switch string(value) {
			case "pending", "in_progress", "completed", "failed":
				decoded.Status = value
			}
		}
	}
	if err := requireJSONFieldsOnly(data, "toolCallId", "title"); err != nil {
		return err
	}
	*c = ToolCall(decoded)
	return nil
}

// ToolCallContent: Content produced by a tool call.
//
// Tool calls can produce different types of content including
// standard content blocks (text, images) or file diffs.
//
// See protocol docs: [Content](https://agentclientprotocol.com/protocol/tool-calls#content)
type ToolCallContent struct {
	Type       ToolCallContentType `json:"type"`
	Meta       Meta                `json:"_meta,omitzero"`
	Content    ContentBlock        `json:"content,omitzero"`
	NewText    string              `json:"newText,omitempty"`
	OldText    *string             `json:"oldText,omitempty"`
	Path       string              `json:"path,omitempty"`
	TerminalID TerminalID          `json:"terminalId,omitempty"`
}

// ToolCallContentType is the discriminator for ToolCallContent variants.
type ToolCallContentType string

const (
	ToolCallContentTypeContent  ToolCallContentType = "content"
	ToolCallContentTypeDiff     ToolCallContentType = "diff"
	ToolCallContentTypeTerminal ToolCallContentType = "terminal"
)

// ContentToolCallContent creates an ToolCallContent variant: Standard content block (text, images, resources).
func ContentToolCallContent(content ContentBlock) ToolCallContent {
	return ToolCallContent{
		Type:    ToolCallContentTypeContent,
		Content: content,
	}
}

// DiffToolCallContent creates an ToolCallContent variant: File modification shown as a diff.
func DiffToolCallContent(path string, newText string) ToolCallContent {
	return ToolCallContent{
		Type:    ToolCallContentTypeDiff,
		NewText: newText,
		Path:    path,
	}
}

// TerminalToolCallContent creates an ToolCallContent variant: Embed a terminal created with `terminal/create` by its id.
//
// The terminal must be added before calling `terminal/release`.
//
// See protocol docs: [Terminal](https://agentclientprotocol.com/protocol/terminals)
func TerminalToolCallContent(terminalID TerminalID) ToolCallContent {
	return ToolCallContent{
		Type:       ToolCallContentTypeTerminal,
		TerminalID: terminalID,
	}
}

// MarshalJSON implements json.Marshaler.
func (c ToolCallContent) MarshalJSON() ([]byte, error) {
	type alias ToolCallContent
	type wire struct {
		*alias
		Content    *ContentBlock `json:"content,omitempty"`
		NewText    *string       `json:"newText,omitempty"`
		Path       *string       `json:"path,omitempty"`
		TerminalID *TerminalID   `json:"terminalId,omitempty"`
	}
	w := wire{alias: (*alias)(&c)}
	if !reflect.ValueOf(c.Content).IsZero() {
		Content := c.Content
		w.Content = &Content
	}
	if !reflect.ValueOf(c.NewText).IsZero() {
		NewText := c.NewText
		w.NewText = &NewText
	}
	if !reflect.ValueOf(c.Path).IsZero() {
		Path := c.Path
		w.Path = &Path
	}
	if !reflect.ValueOf(c.TerminalID).IsZero() {
		terminalID := c.TerminalID
		w.TerminalID = &terminalID
	}
	switch c.Type {
	case ToolCallContentTypeContent:
		Content := c.Content
		w.Content = &Content
	case ToolCallContentTypeDiff:
		NewText := c.NewText
		w.NewText = &NewText
		Path := c.Path
		w.Path = &Path
	case ToolCallContentTypeTerminal:
		terminalID := c.TerminalID
		w.TerminalID = &terminalID
	}
	return json.Marshal(w)
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *ToolCallContent) UnmarshalJSON(data []byte) error {
	type alias ToolCallContent
	decoded := alias{}
	raw := struct {
		OldText json.RawMessage `json:"oldText"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.OldText) > 0 {
		unmarshalDefault(raw.OldText, &decoded.OldText)
	}
	switch decoded.Type {
	case ToolCallContentTypeContent:
		if err := requireJSONFieldsOnly(data, "type", "content"); err != nil {
			return err
		}
	case ToolCallContentTypeDiff:
		if err := requireJSONFieldsOnly(data, "type", "path", "newText"); err != nil {
			return err
		}
	case ToolCallContentTypeTerminal:
		if err := requireJSONFieldsOnly(data, "type", "terminalId"); err != nil {
			return err
		}
	default:
		return invalidDiscriminator("type", decoded.Type)
	}
	*c = ToolCallContent(decoded)
	return nil
}

// ToolCallID: Unique identifier for a tool call within a session.
type ToolCallID string

// ToolCallLocation: A file location being accessed or modified by a tool.
//
// Enables clients to implement "follow-along" features that track
// which files the agent is working with in real-time.
//
// See protocol docs: [Following the Agent](https://agentclientprotocol.com/protocol/tool-calls#following-the-agent)
type ToolCallLocation struct {
	Meta Meta    `json:"_meta,omitzero"`
	Line *uint32 `json:"line,omitempty"`
	Path string  `json:"path"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (l *ToolCallLocation) UnmarshalJSON(data []byte) error {
	type alias ToolCallLocation
	decoded := alias{}
	raw := struct {
		Line json.RawMessage `json:"line"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Line) > 0 {
		unmarshalDefault(raw.Line, &decoded.Line)
	}
	if err := requireJSONFieldsOnly(data, "path"); err != nil {
		return err
	}
	*l = ToolCallLocation(decoded)
	return nil
}

// ToolCallStatus: Execution status of a tool call.
//
// Tool calls progress through different statuses during their lifecycle.
//
// See protocol docs: [Status](https://agentclientprotocol.com/protocol/tool-calls#status)
type ToolCallStatus string

const (
	// ToolCallStatusPending: The tool call hasn't started running yet because the input is either
	// streaming or we're awaiting approval.
	ToolCallStatusPending ToolCallStatus = "pending"
	// ToolCallStatusInProgress: The tool call is currently running.
	ToolCallStatusInProgress ToolCallStatus = "in_progress"
	// ToolCallStatusCompleted: The tool call completed successfully.
	ToolCallStatusCompleted ToolCallStatus = "completed"
	// ToolCallStatusFailed: The tool call failed with an error.
	ToolCallStatusFailed ToolCallStatus = "failed"
)

// ToolCallUpdate: An update to an existing tool call.
//
// Used to report progress and results as tools execute. All fields except
// the tool call ID are optional - only changed fields need to be included.
//
// See protocol docs: [Updating](https://agentclientprotocol.com/protocol/tool-calls#updating)
type ToolCallUpdate struct {
	Meta       Meta                `json:"_meta,omitzero"`
	Content    *[]ToolCallContent  `json:"content,omitempty"`
	Kind       *ToolKind           `json:"kind,omitempty"`
	Locations  *[]ToolCallLocation `json:"locations,omitempty"`
	RawInput   any                 `json:"rawInput,omitempty"`
	RawOutput  any                 `json:"rawOutput,omitempty"`
	Status     *ToolCallStatus     `json:"status,omitempty"`
	Title      *string             `json:"title,omitempty"`
	ToolCallID ToolCallID          `json:"toolCallId"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (u *ToolCallUpdate) UnmarshalJSON(data []byte) error {
	type alias ToolCallUpdate
	decoded := alias{}
	raw := struct {
		Content   json.RawMessage `json:"content"`
		Kind      json.RawMessage `json:"kind"`
		Locations json.RawMessage `json:"locations"`
		RawInput  json.RawMessage `json:"rawInput"`
		RawOutput json.RawMessage `json:"rawOutput"`
		Status    json.RawMessage `json:"status"`
		Title     json.RawMessage `json:"title"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Content) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.Content, &values); err == nil && values != nil {
			items := []ToolCallContent{}
			for _, value := range values {
				var item ToolCallContent
				if err := json.Unmarshal(value, &item); err == nil {
					switch item.Type {
					case ToolCallContentTypeContent, ToolCallContentTypeDiff, ToolCallContentTypeTerminal:
						items = append(items, item)
					}
				}
			}
			decoded.Content = &items
		}
	}
	if len(raw.Kind) > 0 {
		var value ToolKind
		if err := json.Unmarshal(raw.Kind, &value); err == nil {
			switch string(value) {
			case "read", "edit", "delete", "move", "search", "execute", "think", "fetch", "switch_mode", "other":
				decoded.Kind = &value
			}
		}
	}
	if len(raw.Locations) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.Locations, &values); err == nil && values != nil {
			items := []ToolCallLocation{}
			for _, value := range values {
				var item ToolCallLocation
				if err := json.Unmarshal(value, &item); err == nil {
					items = append(items, item)
				}
			}
			decoded.Locations = &items
		}
	}
	if len(raw.RawInput) > 0 {
		unmarshalDefault(raw.RawInput, &decoded.RawInput)
	}
	if len(raw.RawOutput) > 0 {
		unmarshalDefault(raw.RawOutput, &decoded.RawOutput)
	}
	if len(raw.Status) > 0 {
		var value ToolCallStatus
		if err := json.Unmarshal(raw.Status, &value); err == nil {
			switch string(value) {
			case "pending", "in_progress", "completed", "failed":
				decoded.Status = &value
			}
		}
	}
	if len(raw.Title) > 0 {
		unmarshalDefault(raw.Title, &decoded.Title)
	}
	if err := requireJSONFieldsOnly(data, "toolCallId"); err != nil {
		return err
	}
	*u = ToolCallUpdate(decoded)
	return nil
}

// ToolKind: Categories of tools that can be invoked.
//
// Tool kinds help clients choose appropriate icons and optimize how they
// display tool execution progress.
//
// See protocol docs: [Creating](https://agentclientprotocol.com/protocol/tool-calls#creating)
type ToolKind string

const (
	// ToolKindRead: Reading files or data.
	ToolKindRead ToolKind = "read"
	// ToolKindEdit: Modifying files or content.
	ToolKindEdit ToolKind = "edit"
	// ToolKindDelete: Removing files or data.
	ToolKindDelete ToolKind = "delete"
	// ToolKindMove: Moving or renaming files.
	ToolKindMove ToolKind = "move"
	// ToolKindSearch: Searching for information.
	ToolKindSearch ToolKind = "search"
	// ToolKindExecute: Running commands or code.
	ToolKindExecute ToolKind = "execute"
	// ToolKindThink: Internal reasoning or planning.
	ToolKindThink ToolKind = "think"
	// ToolKindFetch: Retrieving external data.
	ToolKindFetch ToolKind = "fetch"
	// ToolKindSwitchMode: Switching the current session mode.
	ToolKindSwitchMode ToolKind = "switch_mode"
	// ToolKindOther: Other tool types (default).
	ToolKindOther ToolKind = "other"
)

// UnstructuredCommandInput: All text that was typed after the command name is provided as input.
type UnstructuredCommandInput struct {
	Meta Meta   `json:"_meta,omitzero"`
	Hint string `json:"hint"`
}

// Usage is token usage for a prompt turn, from @agentclientprotocol/sdk v1.4.0
// schema/schema.json. This extension is unstable and may change or be removed.
// Optional counters use pointers to distinguish unreported usage from zero.
type Usage struct {
	Meta              Meta    `json:"_meta,omitzero"`
	TotalTokens       uint64  `json:"totalTokens"`
	InputTokens       uint64  `json:"inputTokens"`
	OutputTokens      uint64  `json:"outputTokens"`
	ThoughtTokens     *uint64 `json:"thoughtTokens,omitempty"`
	CachedReadTokens  *uint64 `json:"cachedReadTokens,omitempty"`
	CachedWriteTokens *uint64 `json:"cachedWriteTokens,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (u *Usage) UnmarshalJSON(data []byte) error {
	type alias Usage
	decoded := alias{}
	raw := struct {
		ThoughtTokens     json.RawMessage `json:"thoughtTokens"`
		CachedReadTokens  json.RawMessage `json:"cachedReadTokens"`
		CachedWriteTokens json.RawMessage `json:"cachedWriteTokens"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.ThoughtTokens) > 0 {
		unmarshalDefault(raw.ThoughtTokens, &decoded.ThoughtTokens)
	}
	if len(raw.CachedReadTokens) > 0 {
		unmarshalDefault(raw.CachedReadTokens, &decoded.CachedReadTokens)
	}
	if len(raw.CachedWriteTokens) > 0 {
		unmarshalDefault(raw.CachedWriteTokens, &decoded.CachedWriteTokens)
	}
	if err := requireJSONFieldsOnly(data, "totalTokens", "inputTokens", "outputTokens"); err != nil {
		return err
	}
	*u = Usage(decoded)
	return nil
}

// UsageUpdate: Context window and cost update for a session.
type UsageUpdate struct {
	Meta Meta   `json:"_meta,omitzero"`
	Cost *Cost  `json:"cost,omitempty"`
	Size uint64 `json:"size"`
	Used uint64 `json:"used"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (u *UsageUpdate) UnmarshalJSON(data []byte) error {
	type alias UsageUpdate
	decoded := alias{}
	raw := struct {
		Cost json.RawMessage `json:"cost"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Cost) > 0 {
		unmarshalDefault(raw.Cost, &decoded.Cost)
	}
	if err := requireJSONFieldsOnly(data, "used", "size"); err != nil {
		return err
	}
	*u = UsageUpdate(decoded)
	return nil
}

// WaitForTerminalExitRequest: Request to wait for a terminal command to exit.
type WaitForTerminalExitRequest struct {
	Meta       Meta       `json:"_meta,omitzero"`
	SessionID  SessionID  `json:"sessionId"`
	TerminalID TerminalID `json:"terminalId"`
}

// WaitForTerminalExitResponse: Response containing the exit status of a terminal command.
type WaitForTerminalExitResponse struct {
	Meta     Meta    `json:"_meta,omitzero"`
	ExitCode *uint32 `json:"exitCode,omitempty"`
	Signal   *string `json:"signal,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *WaitForTerminalExitResponse) UnmarshalJSON(data []byte) error {
	type alias WaitForTerminalExitResponse
	decoded := alias{}
	raw := struct {
		ExitCode json.RawMessage `json:"exitCode"`
		Signal   json.RawMessage `json:"signal"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.ExitCode) > 0 {
		unmarshalDefault(raw.ExitCode, &decoded.ExitCode)
	}
	if len(raw.Signal) > 0 {
		unmarshalDefault(raw.Signal, &decoded.Signal)
	}
	*r = WaitForTerminalExitResponse(decoded)
	return nil
}

// WriteTextFileRequest: Request to write content to a text file.
//
// Only available if the client supports the `fs.writeTextFile` capability.
type WriteTextFileRequest struct {
	Meta      Meta      `json:"_meta,omitzero"`
	Content   string    `json:"content"`
	Path      string    `json:"path"`
	SessionID SessionID `json:"sessionId"`
}

// WriteTextFileResponse: Response to `fs/write_text_file`
type WriteTextFileResponse struct {
	Meta Meta `json:"_meta,omitzero"`
}

// MarshalJSON implements json.Marshaler.
func (r ListSessionsResponse) MarshalJSON() ([]byte, error) {
	type alias ListSessionsResponse
	a := alias(r)
	if a.Sessions == nil {
		a.Sessions = []SessionInfo{}
	}
	return json.Marshal(a)
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *ListSessionsResponse) UnmarshalJSON(data []byte) error {
	type alias ListSessionsResponse
	decoded := alias{}
	raw := struct {
		NextCursor json.RawMessage `json:"nextCursor"`
		Sessions   json.RawMessage `json:"sessions"`
		*alias
	}{alias: &decoded}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.NextCursor) > 0 {
		unmarshalDefault(raw.NextCursor, &decoded.NextCursor)
	}
	if len(raw.Sessions) > 0 {
		var values []json.RawMessage
		if err := json.Unmarshal(raw.Sessions, &values); err == nil {
			decoded.Sessions = []SessionInfo{}
			for _, value := range values {
				var item SessionInfo
				if err := json.Unmarshal(value, &item); err == nil {
					decoded.Sessions = append(decoded.Sessions, item)
				}
			}
		}
	}
	if err := requireJSONKeysOnly(data, "sessions"); err != nil {
		return err
	}
	*r = ListSessionsResponse(decoded)
	return nil
}
