# Payment Gateway Wrapper

## SQL query to create table

```
CREATE TABLE payment_requests (
    idempotency_key VARCHAR(255) PRIMARY KEY, -- The Guardrail
    user_id         UUID NOT NULL,
    amount          DECIMAL(10, 2) NOT NULL,
    currency        VARCHAR(3) NOT NULL,
    status          VARCHAR(20) NOT NULL, -- STARTED, SUCCEEDED, FAILED
    
    -- Store the response we sent back so we can return it again identically
    response_body   JSONB, 
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```