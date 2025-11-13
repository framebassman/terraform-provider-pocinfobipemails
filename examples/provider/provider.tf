variable "infobip_base_url" {}
variable "infobip_api_key" {}

provider "pocinfobipemails" {
  base_url = var.infobip_base_url
  api_key  = var.infobip_api_key
  # example configuration here
}
