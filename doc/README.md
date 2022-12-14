# Documentation

## Reps-scanner Schema
![](repo-scanner_schema.png)

## Scanning workflow
![](repo-scanner_workflows.png)

## APIs
### API Get repository list
`GET <hostname>:8080/v1/repositories`

Get list of repositories.

**Inputs**

Field | Required | Type | Location | Description
------------- | ------------- | ------------- | ------------- | -------------
**limit** | *(optional)* | integer  | query | Element amount in one page (10 items by default)
**page** | *(optional)* | integer | query | Page offset (1 by default)

**Outputs**

| Result  | Type | Description |
| ------------- | ------------- | ------------- |
| **repository_id** | integer | Repository ID |
| **repository_name** | string | Repository Name |
| **repository_url** | string | Repository Url |
| **is_active** | boolean | `false` is inactive, `true` is active |

**Status**

| Status | Message |
| ------------- | ------------- |
| 200 | Success |
| 400 | Invalid query provided |

**Example**

Request
```bash
$ curl -X GET 'localhost:8080/v1/repositories' \
  -H 'Content-Type: application/json'
```
Response
```json
{
    "status": 200,
    "message": {
        "en": "Success",
        "vn": "Success"
    },
    "data": [
        {
            "repository_id": 4,
            "repository_name": "Blockchain on Go",
            "repository_url": "github.com/trungkh/blockchain-on-go",
            "is_active": true
        },
        {
            "repository_id": 3,
            "repository_name": "JQuery",
            "repository_url": "github.com/jquery/jquery",
            "is_active": true
        }
    ],
    "meta": null
}
```
### API Create new repository
`POST <hostname>:8080/v1/repository`

Create new repository.

**Inputs**

Field | Required | Type | Location | Description
------------- | ------------- | ------------- | ------------- | -------------
**repository_name** | *(required)* | string  | body | Repository Name
**repository_url** | *(required)* | string | body | Repository Url

**Outputs**

| Result  | Type | Description |
| ------------- | ------------- | ------------- |
| **repository_id** | integer | Repository ID |
| **repository_name** | string | Repository Name |
| **repository_url** | string | Repository Url |
| **is_active** | boolean | `false` is inactive, `true` is active |

**Status**

| Status | Message |
| ------------- | ------------- |
| 201 | Success |
| 400 | Invalid payload provided |
| 400 | Invalid url provided |

**Example**

Request
```bash
$ curl -X POST 'localhost:8080/v1/repository' \
  -H 'Content-Type: application/json' \
  -d '{
    "repository_name": "JQuery",
    "repository_url": "https://github.com/jquery/jquery"
  }'
```
Response
```json
{
    "status": 201,
    "message": {
        "en": "Success",
        "vn": "Success"
    },
    "data": [
        {
            "repository_id": 3,
            "repository_name": "JQuery",
            "repository_url": "github.com/jquery/jquery",
            "is_active": true
        }
    ],
    "meta": null
}
```
### API Edit repository
`PUT <hostname>:8080/v1/repository/{repository_id}`

Edit existing repository by given *{repository_id}*.

**Inputs**

Field | Required | Type | Location | Description
------------- | ------------- | ------------- | ------------- | -------------
**repository_id** | *(required)* | integer | path | Repository ID
**repository_name** | *(optional)* | string  | body | Repository Name
**repository_url** | *(optional)* | string | body | Repository Url
**is_active** | *(optional)* | boolean | body | `false` is inactive, `true` is active

**Outputs**

| Result  | Type | Description |
| ------------- | ------------- | ------------- |
| **repository_id** | integer | Repository ID |
| **repository_name** | string | Repository Name |
| **repository_url** | string | Repository Url |
| **is_active** | boolean | `false` is inactive, `true` is active |

**Status**

| Status | Message |
| ------------- | ------------- |
| 200 | Success |
| 400 | Repository not found |
| 400 | Invalid payload provided |
| 400 | Invalid url provided |
| 406 | Nothing to update |

**Example**

Request
```bash
$ curl -X PUT 'localhost:8080/v1/repository/3' \
  -H 'Content-Type: application/json' \
  -d '{
    "is_active": false
  }'
```
Response
```json
{
    "status": 200,
    "message": {
        "en": "Success",
        "vn": "Success"
    },
    "data": [
        {
            "repository_id": 3,
            "repository_name": "JQuery",
            "repository_url": "github.com/jquery/jquery",
            "is_active": false
        }
    ],
    "meta": null
}
```
### API Delete repository
`DEL <hostname>:8080/v1/repository/{repository_id}`

Delete existing repository by given *{repository_id}*.

**Inputs**

Field | Required | Type | Location | Description
------------- | ------------- | ------------- | ------------- | -------------
**repository_id** | *(required)* | integer | path | Repository ID

**Output Status**

| Status | Message |
| ------------- | ------------- |
| 200 | Success |
| 400 | Repository not found |

**Example**

Request
```bash
$ curl -X DEL 'localhost:8080/v1/repository/3' \
  -H 'Content-Type: application/json'
```
Response
```json
{
    "status": 200,
    "message": {
        "en": "Success",
        "vn": "Success"
    },
    "data": null,
    "meta": null
}
```
### API Trigger a scan
`POST <hostname>:8080/v1/repository/{repository_id}/scan`

Trigger a scan by given *{repository_id}*.

**Inputs**

Field | Required | Type | Location | Description
------------- | ------------- | ------------- | ------------- | -------------
**repository_id** | *(required)* | integer  | path | Repository ID

**Outputs**

| Result  | Type | Description |
| ------------- | ------------- | ------------- |
| **scanning_id** | integer | Scanning ID |
| **repository_id** | integer | Repository ID |
| **scanning_status** | string | Scanning Status<br />`queued` is in queue<br />`in_progress` is in progress<br />`success` is successful<br />`failure` is failed |
| **findings** | array[object] | Finding results |
| **queued_at** | timestampt | Queued Time |
| **scanning_at** | timestampt | Scanning Time |
| **finished_at** | timestampt | Finished Time |

**Status**

| Status | Message |
| ------------- | ------------- |
| 201 | Success |
| 400 | Repository not found |
| 400 | Repository is inactive |

**Example**

Request
```bash
$ curl -X POST 'localhost:8080/v1/repository/3/scan' \
  -H 'Content-Type: application/json'
```
Response
```json
{
    "status": 201,
    "message": {
        "en": "Success",
        "vn": "Success"
    },
    "data": {
        "scanning_id": 17,
        "repository_id": 3,
        "findings": {},
        "scanning_status": "queued",
        "queued_at": "2022-11-28T12:02:46.556284Z",
        "scanning_at": null,
        "finished_at": null
    },
    "meta": null
}
```
### API Get scanning results
`GET <hostname>:8080/v1/scanning/result`

Get list of recent results.

**Inputs**

Field | Required | Type | Location | Description
------------- | ------------- | ------------- | ------------- | -------------
**limit** | *(optional)* | integer  | query | Element amount in one page (10 items by default)
**page** | *(optional)* | integer | query | Page offset (1 by default)
**sort** | *(optional)* | string | query | Sort by created time:<br />`asc` - ascending<br />`desc` - descending (by default)
**status** | *(optional)* | string | query | Filter scanning status:<br />`all` - get all (by default)<br />`queued` - in queue<br />`in_progress` - in progress<br />`success` - successful<br />`failure` - failed

**Outputs**

| Result  | Type | Description |
| ------------- | ------------- | ------------- |
| **scanning_id** | integer | Scanning ID |
| **repository_name** | string | Repository Name |
| **repository_url** | string | Repository Url |
| **scanning_status** | string | Scanning Status:<br />`queued` is in queue<br />`in_progress` is in progress<br />`success` is successful<br />`failure` is failed |
| **findings** | array[object] | Finding results |
| **queued_at** | timestampt | Queued Time |
| **scanning_at** | timestampt | Scanning Time |
| **finished_at** | timestampt | Finished Time |

**Status**

| Status | Message |
| ------------- | ------------- |
| 200 | Success |
| 400 | Invalid query provided |

**Example**

Request
```bash
$ curl -X GET 'localhost:8080/v1/scanning/result' \
  -H 'Content-Type: application/json'
```
Response
```json
{
    "status": 200,
    "message": {
        "en": "Success",
        "vn": "Success"
    },
    "data": [
        {
            "scanning_id": 17,
            "repository_name": "JQuery",
            "repository_url": "github.com/jquery/jquery",
            "findings": [
                {
                    "ID": "24d9c5378e6311f75bfb38246683dab15aafbd1feb725f25c26d0ff7fa254170",
                    "Line": 1,
                    "Action": "filename",
                    "Comment": "Can contain credentials for NPM registries",
                    "FileURL": "https://api.github.com/repos/jquery/jquery/blob/main/.npmrc",
                    "FilePath": ".npmrc",
                    "CommitURL": "",
                    "CommitHash": "",
                    "Description": "NPM configuration file",
                    "LineContent": "save-e",
                    "CommitAuthor": "",
                    "CommitMessage": "",
                    "IsTestContext": false,
                    "RepositoryURL": "https://api.github.com/repos/jquery/jquery",
                    "RepositoryName": "jquery",
                    "RepositoryOwner": ""
                }
            ],
            "scanning_status": "success",
            "queued_at": "2022-11-28T12:02:46.556284Z",
            "scanning_at": "2022-11-28T12:02:46.565262Z",
            "finished_at": "2022-11-28T12:02:49.866722Z"
        }
    ],
    "meta": null
}
```

## Some words
+ Repo-scanner needs bellow components:
    + Gin-gonic for web frameworks.
    + Grab <a href="https://github.com/grab/secret-scanner">secrect-scanner</a> based on <a href="https://github.com/michenriksen/gitrob">Gitrob</a>.
    + PostgresSQL database.
+ Designed by 3 main layers:
    + Delivery layer: for input validation and handling response as JSON.
    + Usecase layer: for business logic handling.
    + Repository layer: for database input/output.
    + And other libraries
+ Why do I chose secret-scanner as repository scanner?
    + It can scan directly to github and other git providers like gitlab and bitbucket.
    + It can scan any kind of repo written by any programming language, not scoped in golang.
    + However, I haven't found any way to scan secret word prefix by `public_key`/`private_key`. This could be cons by the library even I think about <a href="https://github.com/securego/gosec">gosec</a> could solve my problem, though. But not enough time.
+ Tried too much for cover UT cases as much as possible, so lacking of significant cases when time up :(.