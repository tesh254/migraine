variables {
    name = "test-interactive"
}

workflow {
    steps [
        {
            cmd = `bash -c 'read -p "Do you want to proceed? (y/n): " answer && echo "You answered: $answer"'`
            desc = "Test boolean input prompt"
        },
        {
            cmd = `bash -c 'read -p "Deploy to production? (y/n): " confirm && if [ "$confirm" = "y" ]; then echo "Deploying..."; else echo "Cancelled."; fi'`
            desc = "Test another yes/no prompt"
        }
    ]
}