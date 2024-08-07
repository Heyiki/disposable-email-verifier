# Disposable Email Verifier

This is a Go-based API for verifying disposable email addresses. The API checks whether an email address belongs to a known disposable email provider.

## Features

- **Disposable Email Check**: Verifies if an email address is from a disposable email provider.
- **Rate Limiting**: Supports daily request limits to prevent abuse.
- **Automatic Domain Updates**: Periodically updates the list of disposable email domains.

## Requirements

- Go 1.18 or higher
- Vercel account for deployment

## Data Sources

https://github.com/disposable/disposable-email-domains

## Configuration

### Environment Variables

`REQUEST_LIMIT`: (Optional) Sets the maximum number of requests allowed per day. Defaults to 20 if not set.

## Usage

### Verify Disposable Email

Send a GET request to the /verify endpoint with an email query parameter:

```
curl "https://your-project-name.vercel.app/verify?email=example@example.com"
```

Replace your-project-name with your Vercel project name.

### Response

The API will respond with a JSON object indicating whether the email address is disposable:

```
{
    "code": 200,
    "data": {
        "disposable": true
    },
    "msg": "success"
}
```

### Test Link

Accessed using a browser

```
https://disposable-email-verifier.vercel.app/verify?email=example@0-180.com
```
