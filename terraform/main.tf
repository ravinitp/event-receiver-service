provider "aws" {
  region = var.region
}

resource "aws_ecs_cluster" "event_receiver_cluster" {
  name = "event-receiver-cluster"
}

resource "aws_iam_role" "ecs_task_execution_role" {
  name = "ecsTaskExecutionRole"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "ecs_task_execution_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role_policy" "s3_access_policy" {
  name = "s3-access-policy"
  role = aws_iam_role.ecs_task_execution_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "s3:PutObject",
          "s3:PutObject"
        ],
        Effect   = "Allow"
        Resource = "arn:aws:s3:::*${var.s3_bucket_name}/*"
      }
    ]
  })
}

resource "aws_ecs_task_definition" "event_receiver_task" {
  family                   = "event-receiver-task"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn

  container_definitions = jsonencode([
    {
      name      = "event-receiver"
      image     = "${var.image}:latest"
      essential = true
      portMappings = [
        {
          containerPort = 8000
          hostPort      = 8080
        }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = "/ecs/event-receiver"
          "awslogs-region"        = var.region
          "awslogs-stream-prefix" = "event-receiver"
        }
      }
    }
  ])
}

resource "aws_ecs_service" "event_receiver_service" {
  name            = "event-receiver-service"
  cluster         = aws_ecs_cluster.event_receiver_cluster.id
  task_definition = aws_ecs_task_definition.event_receiver_task.arn
  desired_count   = 2
  launch_type     = "FARGATE"

  network_configuration {
    assign_public_ip = true
    subnets          = var.subnets
    security_groups  = var.security_groups
  }
}

resource "aws_cloudwatch_log_group" "event_receiver_logs" {
  name = "/ecs/event-receiver"
}