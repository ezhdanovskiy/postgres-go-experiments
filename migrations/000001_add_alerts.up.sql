CREATE TABLE "alerts"
(
    "id"                bigserial PRIMARY KEY,
    "email"             varchar     NOT NULL,
    "symbol"            varchar     NOT NULL,
    "price"             bigint      NOT NULL,
    "marked_to_send_at" timestamptz,
    "sent_at"           timestamptz,
    "created_at"        timestamptz NOT NULL DEFAULT (now()),
    "updated_at"        timestamptz NOT NULL DEFAULT (now()),
    UNIQUE (email, symbol, price)
);

CREATE INDEX ON "alerts" ("email");
CREATE INDEX ON "alerts" ("symbol");


INSERT INTO alerts (email, symbol, price, marked_to_send_at)
VALUES ('elail01@gmail.com', 'BTC', 5000000, now()),
       ('elail01@gmail.com', 'ETH', 300000, now()),
       ('elail01@gmail.com', 'ETH', 300001, now()),
       ('elail01@gmail.com', 'ETH', 300002, now()),
       ('elail01@gmail.com', 'ETH', 300003, now()),
       ('elail01@gmail.com', 'ETH', 300004, now()),
       ('elail01@gmail.com', 'ETH', 300005, now()),
       ('elail01@gmail.com', 'ETH', 300006, now());
