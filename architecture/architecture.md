flowchart LR
    Client[Client / Postman / curl]
    subgraph OS[Order Service]
        OH[HTTP Handler]
        OU[Order Use Case]
        OR[Order Repository]
        OC[Redis Cache]
        ODB[(Order DB)]
    end
    subgraph PS[Payment Service]
        PH[HTTP Handler]
        PU[Payment Use Case]
        PR[Payment Repository]
        PDB[(Payment DB)]
        Pub[RabbitMQ Publisher]
    end
    subgraph MQ[Message Broker]
        Queue[(payment.completed<br/>Durable Queue)]
    end
    subgraph NS[Notification Service]
        Consumer[RabbitMQ Consumer]
        Worker[Background Worker]
        Store[Redis Status Store]
        Provider[EmailSender / NotificationProvider]
    end
    Redis[(Redis)]

    Client -->|POST /orders| OH
    OH --> OU
    OU --> OC
    OC -->|cache hit| OU
    OU -->|cache miss| OR --> ODB
    OR -->|fresh order| OU
    OU -->|write-through / invalidate| OC
    OU -->|gRPC ProcessPayment| PH
    PH --> PU
    PU --> PR --> PDB
    PU -->|after DB commit| Pub
    Pub -->|publish JSON event| Queue
    Queue -->|consume message| Consumer
    Consumer --> Worker
    Worker --> Store --> Redis
    Worker --> Provider
    Provider -->|simulated SMTP or real SMTP| Worker
    Worker -->|retry with exponential backoff| Provider
    Worker -->|manual ACK after success| Consumer
    Client -->|GET /orders/{id}| OH
