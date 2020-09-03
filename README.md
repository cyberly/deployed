# deployed

A simple rollout verification tool to be used in conjunction with [this Azure DevOps task](https://docs.microsoft.com/en-us/azure/devops/pipelines/tasks/utility/http-rest-api?view=azure-devops) in callback mode.  
  
## Configuration
* Move the default JSON headers into the POST body field (the headers can be safely removed)
* Required: Add the following non-Azure DevOps provided values:
  * Namespace - Where to look for deployments
  * Image - The expected image for the deployment
  * Timeout - (optional) The amount of seconds before the listener will stop looking for deployment updates

