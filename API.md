# API 文档

本文档描述 AI 生图后端服务当前已实现的 HTTP 接口。

## 基础信息

- 默认地址：`http://127.0.0.1:3000`
- 请求格式：`application/json`
- 响应格式：`application/json`
- 鉴权方式：`Authorization: Bearer <token>`

错误响应统一格式：

```json
{
  "message": "错误信息"
}
```

## 健康检查

### GET /api/health

用于检查服务是否正常运行。

请求示例：

```bash
curl http://127.0.0.1:3000/api/health
```

成功响应：

```json
{
  "status": "ok"
}
```

## 用户注册

### POST /api/auth/register

用于创建新用户。

请求体：

```json
{
  "phone": "13800138000",
  "username": "demo_user",
  "password": "demo_password"
}
```

字段说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| phone | string | 是 | 手机号，必须唯一 |
| username | string | 是 | 用户账号，必须唯一 |
| password | string | 是 | 用户密码 |

成功响应：

```json
{
  "user": {
    "id": 1,
    "phone": "13800138000",
    "username": "demo_user",
    "createdAt": "2026-07-01T10:00:00+08:00",
    "updatedAt": "2026-07-01T10:00:00+08:00"
  }
}
```

## 用户登录

### POST /api/auth/login

用于登录并获取 JWT Token。`account` 可以传手机号或用户账号。

请求体：

```json
{
  "account": "demo_user",
  "password": "demo_password"
}
```

字段说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| account | string | 是 | 手机号或用户账号 |
| password | string | 是 | 用户密码 |

成功响应：

```json
{
  "token": "jwt-token",
  "user": {
    "id": 1,
    "phone": "13800138000",
    "username": "demo_user",
    "createdAt": "2026-07-01T10:00:00+08:00",
    "updatedAt": "2026-07-01T10:00:00+08:00"
  }
}
```

## 获取当前用户

### GET /api/auth/me

用于校验 Token 并返回当前登录用户。

请求头：

```text
Authorization: Bearer <token>
```

成功响应：

```json
{
  "user": {
    "id": 1,
    "phone": "13800138000",
    "username": "demo_user",
    "createdAt": "2026-07-01T10:00:00+08:00",
    "updatedAt": "2026-07-01T10:00:00+08:00"
  }
}
```

## 生成图片

### POST /api/images/generate

用于提交生图任务。该接口需要登录。

后端会将 `prompt` 转发给环境变量 `IMAGE_API_BASE_URL` 对应的生图服务，并使用环境变量中的 `IMAGE_API_KEY` 和 `IMAGE_API_APP_ID` 调用外部接口。外部接口返回 Base64 或图片 URL 后，后端会保存为 PNG 文件，并将图片记录写入 MySQL。

请求头：

```text
Authorization: Bearer <token>
Content-Type: application/json
```

请求体：

```json
{
  "prompt": "一座未来感城市，夜晚，霓虹灯，电影质感",
  "count": 1
}
```

字段说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| prompt | string | 是 | 生图提示词 |
| count | number | 否 | 生成数量，默认 `1`，最大 `50` |

成功响应：

```json
{
  "items": [
    {
      "id": 1,
      "userId": 1,
      "prompt": "一座未来感城市，夜晚，霓虹灯，电影质感",
      "imagePath": "/storage/images/user_1/20260701_100000_000000000.png",
      "status": "success",
      "durationMs": 5320,
      "createdAt": "2026-07-01T10:00:05+08:00",
      "updatedAt": "2026-07-01T10:00:05+08:00"
    }
  ]
}
```

失败记录也会尝试写入数据库，响应中的 `status` 可能为 `failed`，并带有 `errorMessage`。

## 获取图片集

### GET /api/images

用于获取当前用户的图片生成记录。该接口需要登录。

请求头：

```text
Authorization: Bearer <token>
```

查询参数：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- | --- |
| page | number | 否 | `1` | 页码 |
| pageSize | number | 否 | `20` | 每页数量，最大 `100` |

请求示例：

```bash
curl "http://127.0.0.1:3000/api/images?page=1&pageSize=20" \
  -H "Authorization: Bearer <token>"
```

成功响应：

```json
{
  "items": [
    {
      "id": 1,
      "userId": 1,
      "prompt": "一座未来感城市，夜晚，霓虹灯，电影质感",
      "imagePath": "/storage/images/user_1/20260701_100000_000000000.png",
      "status": "success",
      "durationMs": 5320,
      "createdAt": "2026-07-01T10:00:05+08:00",
      "updatedAt": "2026-07-01T10:00:05+08:00"
    }
  ],
  "page": 1,
  "pageSize": 20,
  "total": 1
}
```

## 生图服务调用说明

后端当前会向外部生图服务发起如下请求：

```text
POST {IMAGE_API_BASE_URL}/v1/images/generations
```

请求头：

```text
Content-Type: application/json
Authorization: Bearer {IMAGE_API_KEY}
X-App-ID: {IMAGE_API_APP_ID}
```

请求体：

```json
{
  "model": "gpt-image-2",
  "prompt": "用户输入的提示词",
  "n": 1,
  "quality": "auto",
  "output_format": "png",
  "size": "1024x1024"
}
```

后端支持解析外部接口响应中的以下图片字段：

- `data[].b64_json`
- `data[].base64`
- `data[].image_b64`
- `data[].url`

当响应为 Base64 时，后端会直接解码保存为 PNG；当响应为 URL 时，后端会下载图片后保存。
