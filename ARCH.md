## Project Structure

This is a full-stack project.

`frontend` is under `front` dir. Using: `vue3`,`primevue`,`tailwindcss`. Version info see `front/package.json`.
Here is `frontend` structure details:
* `front/src/components`: reuseble vue components
* `front/src/assets`: place static image and css file
* `front/src/pocketbase/index.js`: global `pocketbase` client which interact with backend server
* `front/src/router/index.js`: vue router definition
* `front/src/schema`: define rules used by validating on obj saving by using `yup`
* `front/src/util/phone.js`: integrate `jssip` to conmunicate with `freeswitch` through `WSS`
* `front/src/stores`: `pinia` global store
* `front/src/stores/user.js`: `pinia` store, user login/logout and so on
* `front/src/views`: specific vue page
* `front/src/views/sys`: system config
* `front/src/views/num`: phone number related
* `front/src/views/task`: task related

`backend` using: `golang`, `pocketbase`.
* `cmd/lightcall/main.go`: start the backend server
- `server/` - Core business logic
  - `call/` - Call processing and FreeSWITCH integration
  - `config/` - Configuration management
  - `cloud/` - Cloud service integrations
  - `tail/` - Log processing (CDR, CDC)
  - `appender/` - append change data (activity)
* `sql/app/*.go`: `pocketbase` migration files. file begin with `dev-` is only include under development.
* `sql/app/dev-data/*.json`: data used by project development. its filename indicate name of collection created on `pocketbase`.
such as `sql/app/dev-data/users.json`, filename `user` indicate collection `user`. `dev-data` will be load when `backend` doing `migration`.

**You can read `example/dev/dev-data` json file to see the collection structure**


## Data Model (PocketBase Collections)

* `config`: 系统设置, 其中`name`字段为配置名称, `value`是json结构. 每个配置都在`server/config`中有对应的结构体

* `users`: 系统用户
  ```json
  {
    "id": "ddeevvuser00001",
    "avatar": "",
    "email": "li@test.com",
    "emailVisibility": false,
    "name": "小李",
    "password": "123123123",
    "isAdmin": true,
    "active": true,
    "verified": true
  }
  ```

* `outgw`: 外呼网关, 执行实际呼叫时使用.
  ```json
  {
    "id": "t88gsc1c77q0bqe",
    "name": "某讯科技出口",
    "protocal": "SIP",
    "addr": "192.168.66.30:5080",
    "enable": true,
    "options": {
      "password": "432111",
      "registry": false
    },
    "transcaller": [{ "type": "prefix", "param": ["1#"] }],
    "status": {
      "ok": true,
      "error": "",
      "updated": "2024-09-14 13:32"
    }
  }
  ```

* `number`: 外呼号码, 执行实际呼叫时使用.
  ```json
  {
    "id": "16ntm4xzl8c5unk",
    "number": "1232123",
    "outgw": "t88gsc1c77q0bqe",
    "enable": true,
    "mark": {
      "mark": [
        {
          "from": "vivo",
          "type": "疑似诈骗",
          "severity": "danger",
          "cnt": 2
        }
      ],
      "updated": "2024-09-14 13:32"
    },
    "tag": {
      "city": "北京市",
      "province": "北京市"
    }
  }
  ```

* `objective`: 目标, 通常是一个待分解的事情. 在一个目标下有多个`task`. `object` 拥有很多公共信息会共享给它的`task`. 进行到某种程度`object`会被归档. `ext_id` 为同步时用来保存三方系统的标识.`docs` 为该目标的相关资料, 可以上传pdf或者图片等.
  ```json
  {
    "id": "devobjective001",
    "ext_id": "devobjective001",
    "title": "20250314#A公司销售开发",
    "info": {
      "company": "北京智云科技有限公司",
      "background": "背景资料/销售资料........."
    },
    "tasks": ["ddeevvtask00001"],
    "docs": [],
    "open": true
  }
  ```

* `task`: 任务, 具体的一个呼叫对象, 会被分配给一个`user`. `user` 拨打该`task` 会产生`activity`. 进行到一定阶段`task` 会被关闭.`ext_id` 为同步时用来保存三方系统的标识. `ext_id` 和 `id` 一样时为本系统创建的, 不一样时来自三方系统的创建.
  ```json
  {
    "id": "ddeevvtask00001",
    "ext_id": "ddeevvtask00001",
    "own": "ddeevvuser00001",
    "contact": "张经理",
    "callee": "010-1234567",
    "desc": "张经理主管技术;介绍我们的产品尝试销售",
    "activity": ["ddeevvactive001"],
    "open": true
  }
  ```

* `activity`: 活动, 通话时间/时常, 录音, 总结等. `rawlog`为json结构, 记录呼叫过程中的状态. `hook` 为请求云端服务的结果
  ```json
  {
    "id": "ddeevvactive002",
    "user": "ddeevvuser00001",
    "comment": "电话开始于 2024-09-13 13:22, 总用时 5 分钟",
    "record": "are-you-ok.mp3",
    "rawlog": {
          "call": {
              "Addr": "192.168.66.30:5080",
              "Callee": "13500001111321",
              "Caller": "1#1232123",
              "OriCallee": "13500001111",
              "OriCaller": "1232123"
          },
          "cdr": {}
    },
    hook: [ "cloudresp000001" ],
    "isCall": true
  }
  ```

* `cloudresp`: 为调用云端服务的响应
  ```json
  {
      "id": "cloudresp000001",
      "type": "pre-call",
      "name": "blacklist",
      "result": {
          "pass": true,
          "msg": ""
      },
      "rawresp": {
          "code": 0,
          "msg": "success",
          "data": {
              "pass": true
          }
      }
  }
  ```

### Key Relationships
```
objective (1) ----< (N) task (N) ----< (N) activity
                         |
                         v
                      user (owner)

number (N) ----< outgw (1)
```
