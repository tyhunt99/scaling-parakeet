# scaling-parakeet

A simple restful API for managing users.

## Deployment

This service is deployed with the Servless framework.

Assuming you have a valid AWS IAM user created and configuured. To deply to AWS:

```bash
serverless deploy
```

## Testing

To run full tests:

```bash
make test
```

To run tests with coverage output:

```bash
make test-coverage
```
