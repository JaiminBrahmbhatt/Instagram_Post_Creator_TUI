# How to Publish This Project to GitHub

Follow these steps to put **Instagram Post Creator TUI** on GitHub so others can see and clone it.

---

## 1. Create a new repository on GitHub

1. Go to [github.com](https://github.com) and sign in.
2. Click **"+"** (top right) → **"New repository"**.
3. Set:
   - **Repository name:** `Instagram_Post_Creator_TUI` (or any name you like).
   - **Description:** e.g. `TUI app to compose and schedule Instagram posts (Bubble Tea)`.
   - **Public**.
   - Do **not** check "Add a README" or "Add .gitignore" (you already have them).
4. Click **"Create repository"**.

---

## 2. Make this folder its own Git repo and push

Open a terminal in this project folder and run:

```bash
cd /home/maharshi/projects/Instagram_Post_Creator_TUI

# Make this folder a new Git repo (separate from parent)
git init

# Add all files (.env and post_creator.db are ignored by .gitignore)
git add .

# Optional: if you don't want to publish the "my pics" folder, add it to .gitignore first:
#   echo "my pics/" >> .gitignore
# then run: git add .   again

git commit -m "Initial commit: Instagram Post Creator TUI with location, user tags, alt text, stories, share to feed"

# Use the main branch name
git branch -M main

# Add your new GitHub repo as remote (replace YOUR_USERNAME and REPO_NAME with yours)
git remote add origin https://github.com/YOUR_USERNAME/Instagram_Post_Creator_TUI.git

# Push (GitHub will ask for login or token)
git push -u origin main
```

---

## 3. Push your changes to the `dev` branch (after creating `dev` on GitHub)

If you created a branch **dev** on GitHub (from main) and want all your local changes on **dev**:

```bash
cd /home/maharshi/projects/Instagram_Post_Creator_TUI

# 1. Commit everything you have (if not already committed)
git add .
git status   # check what will be committed
git commit -m "Your commit message"   # skip if nothing to commit

# 2. Fetch the dev branch from GitHub
git fetch origin

# 3. Push your current branch (e.g. main) to remote dev — all your commits go to dev
git push origin HEAD:dev
```

**Alternative: switch to dev and push from there**

```bash
# After committing on main:
git fetch origin
git checkout -b dev origin/dev    # create local dev tracking origin/dev (or just: git checkout dev)
git merge main                     # bring all main's commits into dev
git push origin dev
```

---

## 4. Important: never commit secrets

- **`.env`** is already in `.gitignore` — it will **not** be pushed. Keep it that way (it has tokens and app secrets).
- Use **`.env.example`** (no real secrets) so others know which env vars to set.

---

## 5. After the first push

- Your project will be at: `https://github.com/YOUR_USERNAME/Instagram_Post_Creator_TUI`
- Others can clone with:  
  `git clone https://github.com/YOUR_USERNAME/Instagram_Post_Creator_TUI.git`
- To update later:  
  `git add .` → `git commit -m "Your message"` → `git push`

If GitHub asks for a password, use a **Personal Access Token** (Settings → Developer settings → Personal access tokens) instead of your account password.
