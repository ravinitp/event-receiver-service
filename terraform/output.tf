output "ecs_cluster_name" {
  value = aws_ecs_cluster.event_receiver_cluster.name
}

output "ecs_service_name" {
  value = aws_ecs_service.event_receiver_service.name
}