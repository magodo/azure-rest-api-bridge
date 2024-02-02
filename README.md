# azure-rest-api-bridge

A tool to bridge an Azure based application data format (e.g. terraform schema) to its API models.

## Example

Following example will try to run this tool to map terraform-provider-azurerm schema to its API models.

1. Clone the [Azure swagger repo](https://github.com/azure/azure-rest-api-specs), e.g., to `$HOME/github/azure-rest-api-specs`
1. Build up the azure swagger index by using [azure-rest-api-index](https://github.com/magodo/azure-rest-api-index):
    ```shell
    azure-rest-api-index build -o /tmp/index.json $HOME/github/azure-rest-api-specs/specification
    ```

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
    export ARM_PROVIDER_ENHANCED_VALIDATION=1                                                                                                                                                                                       -    
    export ARM_SKIP_PROVIDER_REGISTRATION=true
    ```

1. Prepare the input file for `azure-rest-api-bridge`, which is a HCL file, e.g. `config.hcl`:

    ```hcl
    execution "azurerm_resource_group" "basic" {
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
    azure-rest-api-bridge -port 8888 -config ./config.hcl -index ./index.json -specdir $HOME/github/azure-rest-api-specs/specification
    ```

    It will prints something like below:

    ```
    2023-07-28T16:19:15.743+0800 [INFO]  azure-rest-api-bridge: Starting the mock server
    2023-07-28T16:19:15.743+0800 [INFO]  azure-rest-api-bridge: Executing azurerm_resource_group.basic (1/1)
    {
      "azurerm_resource_group": {
        "/location": [
          {
            "addr": "location",
            "link_github": "https://github.com/Azure/azure-rest-api-specs/blob/fe78d8f1e7bd86c778c7e1cafd52cb0e9fec67ef/specification/resources/resource-manager/Microsoft.Resources/stable/2020-06-01/resources.json#L5439",
            "link_local": "/home/magodo/github/azure-rest-api-specs/specification/resources/resource-manager/Microsoft.Resources/stable/2020-06-01/resources.json:5439:21",
            "ref": "/home/magodo/github/azure-rest-api-specs/specification/resources/resource-manager/Microsoft.Resources/stable/2020-06-01/resources.json#/definitions/ResourceGroup/properties/location"
          }
        ],
        "/managed_by": [
          {
            "addr": "managedBy",
            "link_github": "https://github.com/Azure/azure-rest-api-specs/blob/fe78d8f1e7bd86c778c7e1cafd52cb0e9fec67ef/specification/resources/resource-manager/Microsoft.Resources/stable/2020-06-01/resources.json#L5443",
            "link_local": "/home/magodo/github/azure-rest-api-specs/specification/resources/resource-manager/Microsoft.Resources/stable/2020-06-01/resources.json:5443:22",
            "ref": "/home/magodo/github/azure-rest-api-specs/specification/resources/resource-manager/Microsoft.Resources/stable/2020-06-01/resources.json#/definitions/ResourceGroup/properties/managedBy"
          }
        ],
        "/name": [
          {
            "addr": "name",
            "link_github": "https://github.com/Azure/azure-rest-api-specs/blob/fe78d8f1e7bd86c778c7e1cafd52cb0e9fec67ef/specification/resources/resource-manager/Microsoft.Resources/stable/2020-06-01/resources.json#L5425",
            "link_local": "/home/magodo/github/azure-rest-api-specs/specification/resources/resource-manager/Microsoft.Resources/stable/2020-06-01/resources.json:5425:17",
            "ref": "/home/magodo/github/azure-rest-api-specs/specification/resources/resource-manager/Microsoft.Resources/stable/2020-06-01/resources.json#/definitions/ResourceGroup/properties/name"
          }
        ],
        "/tags/KEY": [
          {
            "addr": "tags/*",
            "link_github": "https://github.com/Azure/azure-rest-api-specs/blob/fe78d8f1e7bd86c778c7e1cafd52cb0e9fec67ef/specification/resources/resource-manager/Microsoft.Resources/stable/2020-06-01/resources.json#L5449",
            "link_local": "/home/magodo/github/azure-rest-api-specs/specification/resources/resource-manager/Microsoft.Resources/stable/2020-06-01/resources.json:5449:35",
            "ref": "/home/magodo/github/azure-rest-api-specs/specification/resources/resource-manager/Microsoft.Resources/stable/2020-06-01/resources.json#/definitions/ResourceGroup/properties/tags/additionalProperties"
          }
        ]
      }
    }
    2023-07-28T16:19:16.585+0800 [INFO]  azure-rest-api-bridge: Stopping the mock server
    ```

## Config Format

The config file is in HCL format, where its basic structure is like below:

```hcl
# 0 or more override blocks that applies for all executions
override {
    #...
}

# 1 or more execution blocks
execution "name" "type" {
    #...
}
```

---

Each `execution` block is defined below:

```hcl
execution "name" "type" {
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

    # 0 or more override blocks, that only applies to this execution
    override {
        #...
    }

    # 0 or more vibrate blocks, that only applies to this execution
    vibrate {
        #...
    }
}
```

The `name` label is used as the key in the resulted output map, while the `type` represents a facets of the `name`. In context of Terraform AzureRM executions, the `name` is used as the resource type, while the `type` is used for different scenarios of the executions.

---

Each `override` block is defined below:

```hcl
override {
    path_pattern = "..." # regexp of the API path pattern, if it is matched against the request sent to the mock server, it will modify the response per response_xxx settings

    request_modify {...}                  # (Optional) A `request_modify` block that can modify the aspects of the request

    # The following ones will conflict
    response_selector_merge = "..."       # (Optional) A JSON object that will be used to select the expected response from multiple synthesized responses.
                                          # It is used as a JSON merge patch, and the first synthesized response that with this patch applied introduce no change will be selected
    resposne_selector_json = "..."        # (Optional) Similar to `response_selector_merge`, but is a JSON patch instead of JSON merge patch.

    # The following ones will conflict
    response_body = "..."                 # (Optional) The response body to return from the mock server for the matched request
    response_patch_merge = "..."          # (Optional) A JSON merge patch that will be applied to the mock response generated by the mock server, for the matched request
    response_patch_json = "..."           # (Optional) A JSON patch that will be applied to the mock response generated by the mock server, for the matched request


    response_header = {                   # (Optional) A map of headers to be returned in the response (e.g. "Content-Type")
        key = value
    }

    response_status_code                  # (Optional) The status code to be returned in the response (otherwise, 200 is returned on success)


    expander    {...}                     # (Optional) A `expander` block that can modify the Swagger expander's behavior
    synthesizer {...}                     # (Optional) A `synthesizer` block that can modify the Swagger synthesizer's behavior
}
```

---

The `request_modify` block is defined below:

```hcl
request_modify {
    method  = "..."      # The request method to change into
    path    = "..."      # The URL path to change into
    version = "..."      # The api-version query parameter to change into
}
```

---

The `expander` block is defined below:

```hcl
expander {
    empty_obj_as_str = false    # Whether to change the schema that is of type "object", but has no other attributes (i.e. properties, additionalProperties, allOf), to be a schema of type "string"?
                                # This is to adpot for poor APIs (e.g. Azure data factory RP) that defines properties as of type `object`, but the API actually returns "string".
    disable_cache    = false    # Whether to disable caching? By default, caching is enabled.
}
```

---

The `synthesizer` block is defined below:

```hcl
synthesizer {
    use_enum_value = false      # Whether to use the defined enum values (pick up the first one)  when synthesizing the response value for the enum properties?

    duplicate_element {...}     # 0 or more `duplicate_element` block that is used to duplicate key/map elements (otherwise, only one element is synthesized).
}
```

---

Each `duplicate_element` block is defined below:

```hcl
duplicate_element {
    count = 1       # (Optional) The count of duplications. By default, 1.
    addr  = "..."   # The address to the property to duplicate, see `PropertyAddr` for the correct format
}
```

---

Each `vibrate` block is defined below:

```hcl
vibrate {
    path_pattern = "..." # regexp of the API path pattern, if it is matched against the request sent to the mock server, it will modify the response per the settings defined in this block
    path = "..."         # The JSON pointer references an *leaf* location within the response (after override), who is of a primary type
    value = "..."        # The value to be applied to above path, which should be different than its original value (the override response)
}
```

---

Note that the execution `name` must be unique.

Available variables:

- `home`: The path to the user's home directory
- `server_addr`: The address in form of "addr:port" that the mock server listens to 

Available functions:

- `jsonencode`
