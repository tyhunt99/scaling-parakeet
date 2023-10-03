# scaling-parakeet

A simple restful API for managing users.

## API Docs

### user

A user is a simple entiry made up of the following attributes.

```javascript
{
    // the uuid of the user
    "id":    string
    // first and last name of the user
    "name":  string
    // email address for the user
    "email"  string
    // birthday of the user
    "dob"    string
    // city where the user is located
    "city"   string
}
```

### GET /users

Returns a list of all users in the database.

### POST /users

Create a user with the specified values. Returns the user with the generated uuid. If an id is set it will be ignored.

```javascript
{
    // Required
    // first and last name of the user
    "name":  string
    // Req: email address for the user
    "email"  string

    // Optional
    // birthday of the user
    "dob"    string
    // city where the user is located
    "city"   string
}
```

### DELETE /users/{id}

Deletes the user with the provided id. It does not return the data on delete.

### GET /users/{id}

Fetches the user with the provided id. Return 404 if the user is not found.

## Development

### Deployment

This service is deployed with the Servless framework.

Assuming you have a valid AWS IAM user created and configuured. To deply to AWS:

```bash
make deploy
```

For a full deploy of the whole stack. not this is potentially destructive and could result in data loss so use caution.

```bash
serverless deploy
```

### Testing

To run full tests:

```bash
make test
```

To run tests with coverage output:

```bash
make test-coverage
```

### Debugging

Deploying the deployed lambda function can be done using

```bash
serverless logs -f handler
```

## Futre Changes

* Validate the user data
* Better testing to actually test logic as opposed to simply testing the mock
* Authentication/authorization
* Helper functions to reduce duplicate code
* Full local support, including a mocked db
* Further instrutions on environment setup
* Add filtering for list
