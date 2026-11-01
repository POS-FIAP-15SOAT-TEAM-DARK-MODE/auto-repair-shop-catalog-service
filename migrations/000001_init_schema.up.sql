CREATE TABLE "work" (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    unit_price NUMERIC(10, 2) NOT NULL,
    status BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE supply (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    unit_price NUMERIC(10, 2) NOT NULL,
    stock_quantity INT NOT NULL DEFAULT 0,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE stock_reservation (
    id               VARCHAR(36) PRIMARY KEY,
    service_order_id VARCHAR(26) NOT NULL,
    supply_id        VARCHAR(36) NOT NULL REFERENCES supply (id),
    quantity         INT NOT NULL CHECK (quantity > 0),
    status           VARCHAR(20) NOT NULL DEFAULT 'RESERVED' CHECK (status IN ('RESERVED', 'RELEASED')),
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    released_at      TIMESTAMP WITH TIME ZONE,
    UNIQUE (service_order_id, supply_id)
);
CREATE INDEX idx_stock_reservation_service_order_id ON stock_reservation (service_order_id);

-- dedupe of consumed messages (message_id = envelope message_id)
CREATE TABLE processed_message (
    message_id   UUID PRIMARY KEY,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
