# comms-proxy

Centralized gateway for sending notifications: callers describe a notification once and name
the channels it goes out on; comms-proxy delivers it on each and hides the vendors behind them.

| Channel  | Status |
|----------|--------|
| In-app   | Live: stored per recipient, read through the Inbox API |
| Push, SMS, WhatsApp, Email | Reserved in the API; reported as "not available yet" |

## APIs

- **Notifier** (`Send`): for other services, over cluster-internal gRPC only (port 8081). It
  has no HTTP route, so it can't be reached through the gateway. Go services use
  [`pkg/client`](pkg/client/client.go).
- **Inbox**: a user's in-app notifications over HTTP, behind the gateway, which supplies the
  `X-Profile-Id` / `X-Profile-Ids` headers.
  - `GET /comms-proxy/v1/notifications?limit=&offset=&unreadOnly=`
  - `GET /comms-proxy/v1/notifications/unread-count`
  - `POST /comms-proxy/v1/notifications/{id}/read`
  - `POST /comms-proxy/v1/notifications/read-all`

Swagger: `/comms-proxy/v1/swagger`.

### Sending

```go
c, _ := client.New(&client.Config{Host: "comms-proxy.zshala-com.svc.cluster.local:8081"})
c.Send(ctx, &commsproxy_v1.SendRequest{
    App:        "zshala",
    Category:   "group.quiz_scheduled",
    Title:      "New quiz in Physics Club",
    Body:       "Kinematics · starts now · 15 min",
    Link:       "/community/quiz?id=…",
    Recipients: []*commsproxy_v1.Recipient{{ProfileId: 7}},
    Channels:   []commsproxy_v1.Channel{commsproxy_v1.Channel_IN_APP},
    DedupKey:   "group_quiz:…", // a recipient gets each key at most once
})
```

Each channel reports how many recipients it sent to, skipped (nothing to reach them by, or a
repeated dedup key) or failed; one channel failing doesn't stop the others.

## Adding a channel

Implement `channels.Channel` (`internal/channels`) and register it in `service.NewService`.
Callers already name channels in their requests, so nothing changes for them.

## Development

```sh
make proto    # regenerate api/ from api/proto (buf)
make migrate  # apply resources/migrations (fill in resources/configs/migrator.yaml)
make run      # reads dev.yaml
make test     # repository tests run against Postgres when COMMS_PROXY_TEST_DSN is set
```

Deployment: pushes to `master` build `ghcr.io/gofreego/comms-proxy` and bump `k8s/`; ArgoCD
syncs `k8s/` into the cluster. The production config comes from configs-prod (see infra's
`comms-proxy-config-secret` role).
