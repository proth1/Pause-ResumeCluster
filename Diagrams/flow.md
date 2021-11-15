```mermaid
sequenceDiagram
EventBridge ->> Lambda: Send Pause Request
Lambda ->>Secrets Manager: Request CastAI API Key
Secrets Manager ->> Lambda: Provide CastAI API Key
Lambda ->> CastAI API: Request list of Clusters
CastAI API ->> Lambda: Respond with list of Clusters
Lambda ->> Lambda: Compare list of clusters with clusters to be paused
Lambda ->> CastAI API: Request cluster be paused/resumed
CastAI API ->> Lambda: Respond with acknowledgement or error

```