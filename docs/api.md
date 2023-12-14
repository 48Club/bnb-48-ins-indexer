# api documentation

## /bnb48_index/v1/balance/list

method POST

request body(json)

```
{
"page" int
"page_size" int (required, 1~256)
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
                "account_id": int,
                "address": string,
                "tick": string
                "tick_hash": string
                "balance": string,
                "create_at": int,
                "update_at": int,
                "delete_at": 0
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
"page" int
"page_size" int (required, 1~256)
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
                "block": int,
                "block_at": int,
                "tx_hash": string,
                "tx_index": int,
                "from": string,
                "to": string,
                "input": string,
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
                "create_at": int,
                "update_at": 0,
                "delete_at": 0
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
