# 📅 IT TOP College Schedule Widget (Backend)

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=flat&logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat&logo=docker&logoColor=white)
![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)

**[English](#-english)** | **[Русский](#-русский)**

---

## 🇬🇧 English

A lightweight Go backend service powering a daily schedule widget for **IT TOP College** students. Built with Go's standard library (`net/http`), PostgreSQL (`pgxpool`), and in-memory per-user caching.

### 🎯 Who is this for?

This project is designed for anyone studying at **IT TOP College** (in Russia or any other country where the college operates) who wants quick, glanceable access to their daily schedule without constantly opening the web journal.

> **Note:** This project was built by a 2nd-year IT TOP College student. There may still be occasional bugs, edge cases, or typos — feel free to debug, fork, open PRs, or customize it however you like!

### ✨ Features

- **Daily Schedule:** Returns your current and upcoming classes for the day (subject, room, start/end time, duration).
- **Exam Alerts:** Automatically checks and displays if you have an exam scheduled for today.
- **Smart Caching:** In-memory per-user caching with automatic TTL refresh and manual refresh cooldowns to avoid spamming the journal API.
- **Zero-Framework Core:** Written purely with Go's `net/http`, structured JSON logging (`slog`), and graceful shutdown.

### 🚀 Quick Start

#### 1. Clone & Configure
```bash
git clone [https://github.com/sqardey/widget.git](https://github.com/sqardey/widget.git)
cd widget
cp .env.example .env
```
Edit `.env` with your port, database connection string, and optional proxy settings.

#### 2. Run with Docker Compose
Host it on any VPS with Docker installed:
```bash
docker compose up -d --build
```

#### 3. Add Your Account to the Database
Insert your IT TOP journal credentials into PostgreSQL and grab your generated API key using `RETURNING key`:
```bash
docker exec -it widget-postgres psql -U postgres -d widget_db -c \
  "INSERT INTO widget_users (login, password) VALUES ('your_login', 'your_password') RETURNING key;"
```

### 📡 API Usage

```http
GET /api/widget?key=<YOUR_API_KEY>&force=0
```

| Parameter | Required | Description |
| :--- | :---: | :--- |
| `key` | **Yes** | Your unique UUID key returned by PostgreSQL. |
| `force` | No | Set to `1` to force a cache refresh from the journal (subject to a 1-hour cooldown). |

#### Example Response
```json
{
  "has_exam": false,
  "current_class": {
    "name_subject": "Software Testing",
    "audience": "304",
    "start_time": "14:00",
    "end_time": "15:30",
    "duration": 90
  },
  "next_class": {
    "name_subject": "Databases",
    "audience": "201",
    "start_time": "15:40",
    "end_time": "17:10",
    "duration": 90
  },
  "updated_at": "14:05"
}
```

### 🎨 Frontend

This repository contains **only the backend**. To display the widget on your phone or desktop (e.g., via iOS **Scriptable**, Android **KWGT**, or a simple web page):
1. Write a lightweight JS client that sends a `GET` request to your hosted server with your `key` (and `force` parameter if needed).
2. Parse the returned JSON object and render it in whatever style you like.

---

##
