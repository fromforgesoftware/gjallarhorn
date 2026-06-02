// Package client is the consumer-facing SDK for Gjallarhorn's gRPC surface. Other
// forge services dial Gjallarhorn and send notifications through it; returned gRPC
// status codes are mapped to kit apierrors.
package client

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"

	apierrors "github.com/fromforgesoftware/go-kit/errors"
	gjallarhornv1 "github.com/fromforgesoftware/gjallarhorn/pkg/api/gjallarhorn/v1"
)

// Notification is the SDK-facing shape of a notification.
type Notification struct {
	ID        string
	RealmID   string
	Recipient string
	Channel   string
	Template  string
	Data      map[string]any
	Subject   string
	Body      string
	Status    string
	LastError string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SendInput carries a notification to deliver. ID and status are
// server-assigned.
type SendInput struct {
	RealmID   string
	Recipient string
	Channel   string
	Template  string
	Data      map[string]any
	Subject   string
	Body      string
}

// Client wraps Gjallarhorn's gRPC surface with kit error mapping.
type Client struct {
	gjallarhorn gjallarhornv1.GjallarhornServiceClient
}

func New(conn grpc.ClientConnInterface) *Client {
	return &Client{gjallarhorn: gjallarhornv1.NewGjallarhornServiceClient(conn)}
}

// NewFromServiceClient is the seam tests use to inject a fake gRPC client.
func NewFromServiceClient(c gjallarhornv1.GjallarhornServiceClient) *Client {
	return &Client{gjallarhorn: c}
}

// Send delivers a notification and returns its persisted state.
func (c *Client) Send(ctx context.Context, in SendInput) (Notification, error) {
	resp, err := c.gjallarhorn.Send(ctx, &gjallarhornv1.SendRequest{
		RealmId:   in.RealmID,
		Recipient: in.Recipient,
		Channel:   in.Channel,
		Template:  in.Template,
		Data:      mapToStruct(in.Data),
		Subject:   in.Subject,
		Body:      in.Body,
	})
	if err != nil {
		return Notification{}, apierrors.FromGRPCError(err)
	}
	return notificationFromProto(resp.GetNotification()), nil
}

// GetNotification returns a notification and its current delivery status.
func (c *Client) GetNotification(ctx context.Context, id string) (Notification, error) {
	resp, err := c.gjallarhorn.GetNotification(ctx, &gjallarhornv1.GetNotificationRequest{Id: id})
	if err != nil {
		return Notification{}, apierrors.FromGRPCError(err)
	}
	return notificationFromProto(resp.GetNotification()), nil
}

func notificationFromProto(n *gjallarhornv1.Notification) Notification {
	out := Notification{
		ID:        n.GetId(),
		RealmID:   n.GetRealmId(),
		Recipient: n.GetRecipient(),
		Channel:   n.GetChannel(),
		Template:  n.GetTemplate(),
		Subject:   n.GetSubject(),
		Body:      n.GetBody(),
		Status:    n.GetStatus(),
		LastError: n.GetLastError(),
	}
	if s := n.GetData(); s != nil {
		out.Data = s.AsMap()
	}
	if ts := n.GetCreatedAt(); ts.IsValid() {
		out.CreatedAt = ts.AsTime()
	}
	if ts := n.GetUpdatedAt(); ts.IsValid() {
		out.UpdatedAt = ts.AsTime()
	}
	return out
}

func mapToStruct(m map[string]any) *structpb.Struct {
	if len(m) == 0 {
		return nil
	}
	s, err := structpb.NewStruct(m)
	if err != nil {
		return nil
	}
	return s
}
