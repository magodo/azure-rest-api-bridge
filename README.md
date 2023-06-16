# azure-rest-api-bridge

A tool to bridge an Azure based application data format (e.g. terraform schema) to its API models.

## Example

Following example will try to run this tool to map terraform-provider-azurerm schema to its API models.

1. Clone the [Azure swagger repo](https://github.com/azure/azure-rest-api-specs), e.g., to `$HOME/github/azure-rest-api-specs`
1. Build up the azure swagger index by using [azure-rest-api-index](https://github.com/magodo/azure-rest-api-index):
    ```shell
    azure-rest-api-index build -o /tmp/index.json -dedup ./azure-rest-api-index/dedup.json $HOME/github/azure-rest-api-specs/specification
    ```

    Note that the `dedup.json` file is maintained by the `azure-rest-api-index` repo, so please clone the repo to get it.

1. Spin up a metadata host by using [azure-metadata-proxy](https://github.com/magodo/azure-metadata-proxy):

    ```shell
    azure-metadata-proxy -port 9999 -metadata='{"resourceManager": "http://localhost:8888", "authentication": {"loginEndpoint": "http://localhost:8888"}}'
    ```

    Note that you must meet the [precondition](https://github.com/magodo/azure-metadata-proxy#precondition) of the tool to make it to be served as a HTTPS server. In above metadata, we redirect request for ARM and Azure login to a local http server at `http://localhost:8888`, which we will spin up later

1. Setup environment for running the tool:

    ```shell
    # This is used by the terraform-provider-azurerm to query endpoints from the azure-metadat-proxy
    #
    export ARM_METADATA_HOSTNAME=localhost:9999

    # This is just to make the provider happy and fast
    #
    export ARM_SUBSCRIPTION_ID=00000000-0000-0000-000000000000
    export ARM_CLIENT_ID=00000000-0000-0000-000000000000
    export ARM_CLIENT_SECRET=123
    export ARM_TENANT_ID=00000000-0000-0000-000000000000
    export ARM_PROVIDER_ENHANCED_VALIDATION=1
    export ARM_SKIP_PROVIDER_REGISTRATION=true
    ```

1. Prepare the input file for `azure-rest-api-bridge`, which is a HCL file, e.g. `config.hcl`:

    ```hcl
    execution "azurerm_resource_group" {
        path = "${home}/go/bin/terraform-client-import"
        args = [
            "-path", 
            "${home}/go/bin/terraform-provider-azurerm",
            "-type",
            "azurerm_resource_group",
            "-id",
            "/subscriptions/foo/resourceGroups/rg",
        ]
    }
    ```

    There can be more than one `execution` blocks, where each one represents a run of some command. Each command expects to print the application model to the stdout if everything works smoothly.

    Note that we are running a tool called [`terraform-client-import`](https://github.com/magodo/terraform-client-go/tree/main/cmd/terraform-client-import), it is a terraform-protol based client, which is a light weighted terraform client only for importing use case.

1. Run the tool:

    ```shell
    azure-rest-api-bridge -port 8888 -config ./config.hcl -index /tmp/index.json -specdir $HOME/github/azure-rest-api-specs/specification
    ```

    It will prints something like below:

    ```
    2023-06-16T18:15:03.131+0800 [INFO]  azure-rest-api-bridge: Starting the mock server
    2023-06-16T18:15:03.131+0800 [INFO]  azure-rest-api-bridge: Executing azurerm_resource_group
    {
      "/location": {
        "addr": "location",
        "link_github": "https://github.com/Azure/azure-rest-api-specs/blob/fda03acb3594cdd152e50146045adcf588b8c6cf/specification/resources/resource-manager/Microsoft.Resources/stable/2020-06-01/resources.json#L5439",
        "link_local": "/home/magodo/github/azure-rest-api-specs/specification/resources/resource-manager/Microsoft.Resources/stable/2020-06-01/resources.json:5439:21",
        "ref": "/home/magodo/github/azure-rest-api-specs/specification/resources/resource-manager/Microsoft.Resources/stable/2020-06-01/resources.json#/definitions/ResourceGroup/properties/location"
      },
      "/name": {
        "addr": "name",
        "link_github": "https://github.com/Azure/azure-rest-api-specs/blob/fda03acb3594cdd152e50146045adcf588b8c6cf/specification/resources/resource-manager/Microsoft.Resources/stable/2020-06-01/resources.json#L5425",
        "link_local": "/home/magodo/github/azure-rest-api-specs/specification/resources/resource-manager/Microsoft.Resources/stable/2020-06-01/resources.json:5425:17",
        "ref": "/home/magodo/github/azure-rest-api-specs/specification/resources/resource-manager/Microsoft.Resources/stable/2020-06-01/resources.json#/definitions/ResourceGroup/properties/name"
      },
      "/tags/KEY": {
        "addr": "tags.*",
        "link_github": "https://github.com/Azure/azure-rest-api-specs/blob/fda03acb3594cdd152e50146045adcf588b8c6cf/specification/resources/resource-manager/Microsoft.Resources/stable/2020-06-01/resources.json#L5449",
        "link_local": "/home/magodo/github/azure-rest-api-specs/specification/resources/resource-manager/Microsoft.Resources/stable/2020-06-01/resources.json:5449:35",
        "ref": "/home/magodo/github/azure-rest-api-specs/specification/resources/resource-manager/Microsoft.Resources/stable/2020-06-01/resources.json#/definitions/ResourceGroup/properties/tags/additionalProperties"
      }
    }
    2023-06-16T18:15:03.601+0800 [INFO]  azure-rest-api-bridge: Stopping the mock server
    ```

## Config Format

The config file is in HCL format, where its basic structure is like below:

```hcl
# 0 or more override blocks that applies for all executions
override {
    #...
}

# 1 or more execution blocks
execution "name" {
    #...
}
```

Each `override` block is defined below:

```hcl
override {
    path_pattern = "..." # regexp of the API path pattern, if it is matched against the request sent to the mock server, it will modify the response per response_xxx settings

    # Exactly one of below attributes must be specified
    #
    response_body = "..." # The response body to return from the mock server for the matched request
    response_merge_patch = "..." # A JSON merge patch that will be applied to the mock response generated by the mock server, for the matched request
    response_json_patch = "..." # A JSON patch that will be applied to the mock response generated by the mock server, for the matched request
}
```

Each `execution` block is defined below:

```hcl
execution "name" {
    # 0 or more override blocks, that only applies to this execution
    override {
        #...
    }

    # additional environment variables for this execution
    env = {
        #...
    }

    # the working directory for this execution 
    dir = "..."

    # the path to the executable that is expected to print the application model as a JSON object to stdout when runs successfully 
    path = "..."

    # the arguments to the executable
    args = [
        #...
    ]
}
```

Available variables:

- `home`: The path to the user's home directory
- `server_addr`: The address in form of "addr:port" that the mock server listens to 

Available functions:

- `jsonencode`
