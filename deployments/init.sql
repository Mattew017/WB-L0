CREATE TABLE IF NOT EXISTS orders (
                        order_uid VARCHAR(50) PRIMARY KEY,
                        track_number VARCHAR(50) NOT null UNIQUE,
                        entry VARCHAR(10) NOT NULL,
                        locale VARCHAR(10) NOT NULL,
                        internal_signature VARCHAR(100),
                        customer_id VARCHAR(50) NOT NULL,
                        delivery_service VARCHAR(50) NOT NULL,
                        shardkey VARCHAR(10) NOT NULL,
                        sm_id INTEGER NOT NULL,
                        date_created TIMESTAMP WITH TIME ZONE NOT NULL,
                        oof_shard VARCHAR(10) NOT NULL
);

CREATE TABLE IF NOT EXISTS deliveries (
                            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                            order_uid VARCHAR(50) NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
                            name VARCHAR(100) NOT NULL,
                            phone VARCHAR(20) NOT NULL,
                            zip VARCHAR(20) NOT NULL,
                            city VARCHAR(100) NOT NULL,
                            address VARCHAR(200) NOT NULL,
                            region VARCHAR(100) NOT NULL,
                            email VARCHAR(100) NOT NULL
);

CREATE TABLE IF NOT EXISTS payments (
                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                          transaction VARCHAR(50) NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
                          request_id VARCHAR(50),
                          currency VARCHAR(10) NOT NULL,
                          provider VARCHAR(50) NOT NULL,
                          amount INTEGER NOT NULL,
                          payment_dt BIGINT NOT NULL,
                          bank VARCHAR(50) NOT NULL,
                          delivery_cost INTEGER NOT NULL,
                          goods_total INTEGER NOT NULL,
                          custom_fee INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS items (
                       rid VARCHAR(50) PRIMARY KEY,
                       track_number VARCHAR(50) NOT NULL REFERENCES orders(track_number),
                       chrt_id INTEGER NOT NULL,
                       price INTEGER NOT NULL,
                       name VARCHAR(100) NOT NULL,
                       sale INTEGER NOT NULL,
                       size VARCHAR(10) NOT NULL,
                       total_price INTEGER NOT NULL,
                       nm_id INTEGER NOT NULL,
                       brand VARCHAR(100) NOT NULL,
                       status INTEGER NOT NULL
);

CREATE INDEX idx_orders_track_number ON orders(track_number);
CREATE INDEX idx_items_track_number ON items(track_number);