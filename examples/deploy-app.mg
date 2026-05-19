# Migraine workflow: Deploy a web application
# This workflow pre-checks environment, builds, deploys, and notifies on success/failure

metadata {
    name = "deploy-app"
    desc = "Build and deploy the web application to staging"
}

variables {
    app_name = "args:APP_NAME"
    env = "args:ENV"
    deploy_host = "env:DEPLOY_HOST"
    slack_webhook = "vault:SLACK_WEBHOOK"
    build_timeout = 300
}

workflow {
    pre_checks [
        {
            cmd = `docker info`
            desc = "Verify Docker daemon is running"
            on_fail = "action:notify_failure"
        },
        {
            cmd = `git diff --quiet HEAD`
            desc = "Ensure working tree is clean"
            on_fail = "action:notify_dirty_tree"
        }
    ]

    steps [
        {
            cmd = `docker build -t {{app_name}}:{{env}} .`
            desc = "Build the Docker image"
            on_fail = "action:notify_failure"
        },
        {
            cmd = `docker push {{app_name}}:{{env}}`
            desc = "Push image to registry"
            on_fail = "action:notify_failure"
        },
        {
            cmd = `ssh deploy@{{deploy_host}} "kubectl set image deployment/{{app_name}} {{app_name}}={{app_name}}:{{env}}"`
            desc = "Roll out the new image on the cluster"
            on_success = "action:notify_success"
            on_fail = "action:rollback"
        }
    ]

    actions {
        notify_failure {
            cmd = `curl -X POST -H 'Content-type: application/json' --data '{"text":"Deploy of {{app_name}} to {{env}} failed!"}' {{slack_webhook}}`
            desc = "Post failure alert to Slack"
        },
        notify_success {
            cmd = `curl -X POST -H 'Content-type: application/json' --data '{"text":"Deploy of {{app_name}} to {{env}} succeeded!"}' {{slack_webhook}}`
            desc = "Post success message to Slack"
        },
        notify_dirty_tree {
            cmd = `git status --short`
            desc = "Show dirty files in working tree"
        },
        rollback {
            cmd = `ssh deploy@{{deploy_host}} "kubectl rollout undo deployment/{{app_name}}"`
            desc = "Roll back to previous deployment"
            on_fail = "action:notify_failure"
        }
    }
}

config {
    store_variables = true
    store_logs = true
    background = false
    global = false
}