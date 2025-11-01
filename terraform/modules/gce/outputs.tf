output "vm_instance_ids" {
  value       = values(google_compute_instance.vm_instances)[*].id
  description = "The IDs of Google Cloud Compute Engine instances"
}
