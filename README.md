# lgtm-rbac-proxy


## Env vars

| Environment Variable  | Description                                                                                                                                                                                                                                                                                             | Required |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- |
| `DB_PATH`             | Specifies the path to the database file. The application will use this path to read from and write to the database.                                                                                                                                                                                     | Yes      |
| `ADMIN_USER_PASSWORD` | Sets the password for the admin user. This password is used in basic authentication to protect certain endpoints. It's important to set a strong password to ensure the security of these endpoints.                                                                                                    | Yes      |
| `CONFIG_PATH`         | Specifies the path to the configuration file. The application will read this file on startup to configure its settings. The file should be in a format that the application can parse, such as JSON or YAML. The exact settings that can be configured depend on the implementation of the application. | Yes      |


## Config File
| Key                      | Description                                                                                                                                                           | Required    |
| ------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------- |
| `logType`                | Specifies the format of the logs. Possible values are `json` for JSON format and `text` for plain text format.                                                        | Yes         |
| `lokiUrl`                | Specifies the URL of the Loki server. The application will send logs to this server.                                                                                  | Yes         |
| `adminUser.username`     | Specifies the username for the admin user. This username is used in basic authentication to protect certain endpoints.                                                | Yes         |
| `adminUser.password`     | This key is not used. The password for the admin user is set via the `ADMIN_USER_PASSWORD` environment variable.                                                      | No          |
| `lokiBasicAuth.enabled`  | Specifies whether basic authentication is enabled for the Loki server. If this is `true`, the `lokiBasicAuth.username` and `lokiBasicAuth.password` keys must be set. | Yes         |
| `lokiBasicAuth.username` | Specifies the username for basic authentication with the Loki server. This key is required if `lokiBasicAuth.enabled` is `true`.                                      | Conditional |
| `lokiBasicAuth.password` | Specifies the password for basic authentication with the Loki server. This key is required if `lokiBasicAuth.enabled` is `true`.                                      | Conditional |

```yaml

logType: json
lokiUrl:  http://localhost:8081
adminUser:
  username: admin
  // Passsword Set via env variable

lokiBasicAuth:
  enabled: true
  username: loki
  password: loki

```



### Endpoints

---

#### Create Token

- **Method:** POST
- **URL:** `/create`
- **Authentication:** Basic Auth

**Body:**

```json
{
    "name": "test-new-type-aba",
    "tenantIds": ["tenant1", "tenant2"],
    "allowedLabels": {
        "MustInclude": "{container=\"grafana-agent\"}",
        "MustExclude": "{cluster=\"k8s-abctest\"}"
    }
}
```

**Response:**

```json
{
    "Token": "glrbac_abc124567"
}
```

---

#### List Tokens

- **Method:** GET
- **URL:** `/tokens`
- **Authentication:** Basic Auth

**Response:**

```json
[
    {
        "Name": "test-new-type-abva",
        "Token": "hashedToken",
        "TenantIds": [
            "tenant1",
            "tenant2"
        ],
        "AllowedLabels": {
            "MustInclude": "{container=\"grafana-agent\"}"
        }
    }
]
```

---

#### Query Loki Logs

- **Method:** GET
- **URL:** `/loki/*`
- **Authentication:** Token Auth (Bearer)

**Response:**

Response from Loki if successful.

---
