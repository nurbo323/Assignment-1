flowchart LR
    Client[Client / Postman / curl]

    subgraph OS[Order Service]
        OH[HTTP Handler]
        OU[Order Use Case]
        OR[PostgreSQL Repository]
        OC[(Redis Cache)]
        ODB[(Order DB)]
    end

    subgraph PS[Payment Service]
        PH[HTTP Handler]
        PU[Payment Use Case]
        PR[PostgreSQL Repository]
        PDB[(Payment DB)]
        Pub[RabbitMQ Publisher]
    end

    subgraph MQ[Message Broker]
        Queue[(payment.completed\nDurable Queue)]
    end

    subgraph NS[Notification Service]
        Consumer[RabbitMQ Consumer]
        Worker[Background Worker]
        Store[(Redis Status Store)]
        Provider[Notification Provider]
    end

    Redis[(Redis)]

    Client -->|POST /orders| OH
    Client -->|GET /orders/{id}| OH
    Client -->|PATCH /orders/{id}/cancel| OH

    OH --> OU
    OU -->|cache lookup| OC
    OC -->|cache hit| OU
    OU -->|cache miss| OR
    OR --> ODB
    OR -->|refresh cache| OC
    OU -->|payment call with 2s timeout| PH

    PH --> PU
    PU --> PR
    PR --> PDB
    PU -->|publish after commit| Pub
    Pub -->|JSON event| Queue

    Queue -->|consume| Consumer
    Consumer --> Worker
    Worker -->|idempotency / lock| Store
    Store --> Redis
    Worker --> Provider
    Provider -->|SIMULATED or REAL SMTP| Worker
    Worker -->|Ack after success| Consumer
