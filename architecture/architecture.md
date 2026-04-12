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
    end

    Client -->|POST /orders| OH
    OH --> OU
    OU --> OR --> ODB
    OU -->|REST POST /payments\nhttp.Client timeout 2s| PH
    PH --> PU --> PR --> PDB
    PH --> OU
    Client -->|GET /orders/{id}| OH
    Client -->|PATCH /orders/{id}/cancel| OH
    Client -->|GET /payments/{order_id}| PH
