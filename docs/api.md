# api documentation

## /bnb48_index/v1/balance/list

method POST

request body(json)

```
{
"page" int
"page_size" int (required, 1~256)
"tick_hash" string (required)
}
```

response(json) (sorted by balance desc)

```
{
    "code": 0,
    "msg": "ok",
    "data": {
        "count": int
        "page": int
        "page_size": int
        "list": [
            {
                "id": int,
                "account_id": int,
                "address": string,
                "tick": string
                "tick_hash": string
                "balance": string,
                "create_at": int,
                "update_at": int,
                "delete_at": int
            }
        ]
    }
}
```

example:

```
curl --location 'http://hostname:port/bnb48_index/v1/balance/list' \
--header 'Content-Type: application/json' \
--data '{
    "page": 0,
    "page_size": 3,
    "tick_hash": "0xd893ca77b3122cb6c480da7f8a12cb82e19542076f5895f21446258dc473a7c2"
}'
```

## /bnb48_index/v1/record/list

method POST

request body(json)

```
{
"tick_hash" string
"page" int
"page_size" int (required, 1~256)
}
```

response(json) (sorted by block_at desc, tx_index desc)

```
{
    "code": 0,
    "msg": "ok",
    "data": {
        "count": int
        "page": int
        "page_size": int
        "list": [
            {
                "id": int,
                "block": int,
                "block_at": int,
                "is_pending": bool,
                "tx_hash": string,
                "tx_index": int,
                "from": string,
                "to": string,
                "input": string,
                "input_decode": {
                    "amt": string,
                    "decimals", int,
                    "lim": string,
                    "max": string,
                    "miners": string | null,
                    "op": string,
                    "p": string,
                    "tick": string,
                    "tick_hash": string,
                    "to": string,
                },
                "type": int,
                "create_at": int,
                "update_at": int,
                "delete_at": int
            }
        ]
    }
}
```

example:

```
curl --location 'http://hostname:port/bnb48_index/v1/record/list' \
--header 'Content-Type: application/json' \
--data '{"page": 0, "page_size":3}'
```

## /bnb48_index/v1/inscription/list

method POST

request body(json)

```
{
"page" int
"page_size" int (required, 1~256)
"protocol" string
"tick_hash" string
"tick" string
"deploy_by" string
"status" (0:all:, 1:in_progress, 2:completed)
}
```

response(json)

```
{
    "code": 0,
    "msg": "ok",
    "data": {
        "count": int
        "page": int
        "page_size": int
        "list": [
            {
                "id": int,
                "tick": string,
                "tick_hash": string,
                "tx_index": int,
                "block": int,
                "block_at": int,
                "decimals": int,
                "max": int,
                "lim": string,
                "miners": string,
                "minted": int,
                "status" : int, #1: in_progress 2: completed
                "holders" : int,
                "deploy_by" : string,
                "protocol": string,
                "create_at": int,
                "update_at": int,
                "delete_at": int
            }
        ]
    }
}
```

example:

```
curl --location 'http://hostname:port/bnb48_index/v1/inscription/list' \
--header 'Content-Type: application/json' \
--data '{"page": 0, "page_size":3}'
```

## /bnb48_index/v1/account/balance

method POST

request body(json)

```
{
"tick_hash" []string
"address" string (required)
}
```

response(json)

```
{
    "code": 0,
    "msg": "ok",
    "data": {
        "wallet": [
            {
                "id": int,
                "account_id": int,
                "address": string,
                "tick": string,
                "tick_hash": string,
                "balance": string,
                "changes": [
                    {
                        "id": int,
                        "block": int,
                        "block_at": int,
                        "is_pending": bool,
                        "tx_hash": string,
                        "tx_index": int,
                        "from": string,
                        "to": string,
                        "input": string,
                        "input_decode": {
                            "amt": string,
                            "decimals", int,
                            "lim": string,
                            "max": string,
                            "miners": string | null,
                            "op": string,
                            "p": string,
                            "tick": string,
                            "tick_hash": string,
                            "to": string,
                        },
                        "type": int,
                        "create_at": int,
                        "update_at": int,
                        "delete_at": int
                    }
                ]
                "decimals": int,
                "create_at": int,
                "update_at": int,
                "delete_at": int
            }
        ]
    }
}
```

example:

```
curl --location 'http://hostname:port/bnb48_index/v1/account/balance' \
--header 'Content-Type: application/json' \
--data '{"tick_hash":"aaaa", "address":"xxx"}'
```

## /bnb48_index/v1/record
method POST

request body(json)

```
{
"tx_hash" string
}
```

response(json) (sorted by op_index desc)

```
{
    "code": 0,
    "msg": "ok",
    "data": {
        "list": [
            {
                "id": int,
                "block": int,
                "block_at": int,
                "tx_hash": string,
                "tx_index": int,
                "op_index": int,
                "from": string,
                "to": string,
                "input": string,
                "input_decode": {
                    "amt": string,
                    "decimals", int,
                    "lim": string,
                    "max": string,
                    "miners": string | null,
                    "op": string,
                    "p": string,
                    "tick": string,
                    "tick_hash": string,
                    "to": string,
                },
                "type": int, // 1：mint 2: transfer
                "create_at": int,
                "update_at": int,
                "delete_at": int
            }
        ]
    }
}
```

example:

```
curl --location 'http://hostname:port/bnb48_index/v1/record' \
--header 'Content-Type: application/json' \
--data '{"tx_hash":"0x5becb0c3599448df8fe5c9cb4820f971858aa2d4254c6a19bf4a64be468e93e0"}'
```
