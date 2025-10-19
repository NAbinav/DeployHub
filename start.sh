#!/bin/bash
set -e

export PORT=${PORT:-8080}
echo "🚀 Starting app on port $PORT"

# Detect environment
if [ -f "package.json" ]; then
    echo "Detected Node.js project"
    if grep -q "next" package.json; then
        echo "Framework: Next.js"
        npm install
        npm run build
        npx next start -p $PORT
    elif grep -q "vite" package.json; then
        echo "Framework: Vite"
        npm install
        npm run build
        npx serve -s dist -l $PORT
    elif grep -q "react" package.json; then
        echo "Framework: React"
        npm install
        npm run build
        npx serve -s build -l $PORT
    else
        echo "Framework: Express or Node app"
        npm install
        npm start
    fi

elif [ -f "requirements.txt" ]; then
    echo "Detected Python project"
    pip install -r requirements.txt
    if grep -q "fastapi" requirements.txt; then
        exec uvicorn main:app --host 0.0.0.0 --port $PORT
    elif grep -q "flask" requirements.txt; then
        export FLASK_RUN_PORT=$PORT
        export FLASK_RUN_HOST=0.0.0.0
        exec flask run
    else
        exec python main.py
    fi

elif [ -f "go.mod" ]; then
    echo "Detected Go project"
    go build -o app .
    ./app

else
    echo "⚠️ Unknown project type — running default command"
    exec "$@"
fi

