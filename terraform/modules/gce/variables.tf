variable "project_id" {
  type        = string
  description = "The ID for your GCP project"
}

variable "region" {
  type        = string
  description = "The region for your GCP project"
}

variable "zone" {
  type        = string
  description = "The region/location for your GCP project"
}

variable "machine_type" {
  type        = string
  description = "The machine type for your GCP project"
}

variable "vm_name" {
  type        = string
  description = "The name of the VM"
}

variable "with_stop_schedule" {
  type        = bool
  description = "Add resources to the GCP project"
  default     = false
}
