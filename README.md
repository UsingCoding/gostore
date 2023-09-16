# gostore - secret store manager

## Conception

gostore use concept of store. 
Each store keeps files in fs with specific implementation (Git).
gostore does not encrypt store or secrets structure, only values.

## Build

```shell
go build -v -o ./bin/gostore ./cmd/gostore
```

## Usage

### Create new store

```shell
gostore init --id mystore

Generated keys:
Public key: age1ejrt99ns0e8zgplhm7zfuppd3dg6yg4ersyzcgtjp0enpcfshfxqgqkgfw
Private key: <KEY>
```

### Add to store

```shell
cat secret-file | gostore add mysite/secret-file
```

### Get secret from store

```shell
gostore get mysite/secret-file
```

### Composite secrets

Each secret may contain several keys

```shell
echo "admin" | gostore add mysite/admin user
echo "1234" | gostore add mysite/admin pass
```

Get secret
```shell
gostore get mysite/admin

pass: 1234
user: admin
```

Get concrete key:
```shell
gostore get mysite/admin pass

pass: 1234
```

### Remove secrets from store

Remove secret:
```shell
gostore rm mysite/admin
```

Remove key from secret:
```shell
gostore rm mysite/admin pass
```

If secret empty after key deletion, secret will be removed 

### List secrets in store

List all secrets
```shell
gostore ls

mystore
└── mysite
    └── admin
```

List secrets subtree

```shell
gostore ls mysite

mysite
└── admin
```


### List stores

```shell
gostore stores list

mystore: /home/user/.gostore/mystore
github: /home/user/.gostore/github
```

### Use other store

```shell
gostore use github
```