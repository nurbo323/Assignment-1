flowchart LR
    Client[Client / Postman / curl]
    subgraph OS[Order Service]
        OH[HTTP Handler]
        OU[Order Use Case]
        OR[Order Repository]
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
        IdemCheck[Idempotency Check<br/>in-memory map]
        Logger[Log Handler]
    end

    Client -->|POST /orders| OH
    OH --> OU
    OU --> OR --> ODB
    OU -->|gRPC ProcessPayment| PH
    PH --> PU
    PU --> PR --> PDB
    PU -->|after DB commit| Pub
    Pub -->|publish JSON event| Queue
    Queue -->|consume message| Consumer
    Consumer --> IdemCheck
    IdemCheck -->|manual ACK| Consumer
    IdemCheck -->|log notification| Logger
    Client -->|GET /orders/{id}| OH
