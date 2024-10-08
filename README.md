[![Build](https://github.com/mbevc1/aws-utils/actions/workflows/build.yaml/badge.svg)](https://github.com/mbevc1/aws-utils/actions/workflows/build.yaml)

# aws-utils
AWS Utils CLI - making complex tasks simpler and quicker

This is a simple CLI tool to help with some common AWS tasks. It's aiming to
simplify and make some taks quicker by abstracting underlying steps or complexity.

e.g. deleting an AWS account is a single step and includes both: terminating AWS
Control Tower Catalog product and closing the account itself.

## Installing

1. Download `aws-utils` from the [releases](https://github.com/mbevc1/aws-utils/releases)
2. Run `aws-utils -v` to check if it's working correctly.
3. Enjoy!

## Usage

Simply run the binary like:

```shell
# aws-utils lz desc

Using region: eu-west-1
-------------------------------
List of deployed Landing Zones:
-------------------------------
ARN: arn:aws:controltower:eu-west-1:123456789012:landingzone/12JZIC8A68Y3AAAA
Version: 3.3
LatestAvailableVersion: 3.3
Manifest:
{
  "accessManagement": {
    "enabled": true
  },
  "centralizedLogging": {
    "accountId": "123456789012",
    "configurations": {
      "accessLoggingBucket": {
        "retentionDays": "3650"
      },
      "loggingBucket": {
        "retentionDays": "365"
      }
    },
    "enabled": true
  },
  "governedRegions": [
    "eu-west-1"
  ],
  "organizationStructure": {
    "sandbox": {
      "name": "Custom"
    },
    "security": {
      "name": "Core"
    }
  },
  "securityRoles": {
    "accountId": "123456789013"
  }
}
Status: ACTIVE
DriftStatus: IN_SYNC
```

## Building

If you want to build your own binarly locally, you can do that by running:

```shell
make build
```

Which should produce a locally binary to run. You'll need Golang compiler.

To run tests there is a Makefile target for that as well:

```shell
make test
```

## Contributing

Report issues/questions/feature requests on in the [issues](https://github.com/mbevc1/aws-utils/issues/new) section.

Full contributing [guidelines are covered here](.github/CONTRIBUTING.md).

## Authors

* [Marko Bevc](https://github.com/mbevc1)
* Full [contributors list](https://github.com/mbevc1/aws-utils/graphs/contributors)

## License

MPL-2.0 Licensed. See [LICENSE](LICENSE) for full details.
<!-- https://choosealicense.com/licenses/ -->
