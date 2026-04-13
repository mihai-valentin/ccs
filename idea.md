# Idea

Built-in Claude Code session management tools are poor and ackward DX/UX wise. I want to create a CLI tool that will allow me to manage CC sessions in an easy way.

The tool must:
1. include list, search (smart search - by session name, root dir, last message, etc), delete session action
2. give a more rabust session tagging/labelling options (also, must be searchable)
3. give a per-project / per-tag groupping (+ by project search) option
4. must be runnable as daemon with an option to open desired session in a new terminal
5. tool must open any session in two steps: cd into root dir, resume CC session by id/label/name

## LLM TODOs

1. Analyze requirements
2. Enrich it with any useful features/updates
3. Plan the imeplementation
4. Split the plan into Shawarma sub tasks (see the shawarma skill)
5. Initi GIT repo with a git ignore file
6. Run the tasks implementation with Shawarma
7. Validate implementation - create bug tasks for every found isuse and feed it to Shawarma; repeat validation after bug fix.
8. Enrich final repo with docs
9. Print usage instructions
10. Give your feedback on how it was to work with Shawarma

