# 🪄 Encanto

[![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Svelte](https://img.shields.io/badge/Svelte_5-FF3E00?style=for-the-badge&logo=svelte&logoColor=white)](https://svelte.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-DC382D?style=for-the-badge&logo=redis&logoColor=white)](https://redis.io/)

**Encanto** is a professional-grade, team-based messaging operations platform designed for scale, reliability, and precision. It streamlines live conversations, team collaboration, and operational analytics into a unified, high-performance workspace.

---

## 🎯 Vision

To provide teams with an industrial-strength communication hub where every message is traceable, every assignment is clear, and every operational detail is governed by robust business logic. Encanto moves beyond simple chat into the realm of **Operational Intelligence**.

---

## ✨ Key Modules & Features

### 📨 Operational Inbox
*   **Real-time Collaboration**: Live-updating conversation threads with assignment controls.
*   **Team Workspace**: Personal and shared tabs (`Assigned`, `Pending`, `Unclaimed`).
*   **Internal Notes**: Collaborate behind the scenes with private notes and mentions.
*   **Rich Timeline**: A complete history of operational events (`conversation_events`) for every chat.

### ⚡ Messaging Engine
*   **Text & Media**: Seamless handling of text, images, and documents.
*   **Typing Simulation**: Human-like message delivery for outward messages.
*   **Delivery Tracking**: Granular tracking from `pending` to `sent` and `failed`.
*   **Failure Recovery**: Integrated retry and revoke mechanisms.

### 🛡️ Governance & Control
*   **Action-Based Permissions**: Fine-grained access control beyond simple roles.
*   **Visibility Layers**: Restricted scope based on organization, team, or explicit allowance.
*   **Licensing Enforcement**: Real-time quota monitoring and HWID-locked license activation.
*   **Safe Degradation**: Automated "Cleanup Mode" to handle usage overages gracefully.

### 📊 Intelligence & Growth
*   **Agent Analytics**: KPI tracking derived from explicit operational facts.
*   **Campaign Management**: Precise broadcast tools with recipient-level outcome tracking.
*   **Audit Logging**: Every sensitive action is recorded for compliance and review.
*   **Background processing**: Reliable handling of long-running tasks and scheduled runs.

---

## 🛠️ Technology Stack

| Layer | Technologies |
| :--- | :--- |
| **Backend** | Go (Golang), Chi Router, pgx, sqlc, WebSockets |
| **Frontend** | Svelte 5, SvelteKit, TypeScript |
| **Database** | PostgreSQL |
| **Cache & State** | Redis |
| **Testing** | Playwright (E2E) |

---

## 🏗️ Project Architecture

Encanto follows a strict **layered architecture** to ensure long-term maintainability:

```text
/
├── frontend/             # Interface: Svelte 5 Web Application
├── backend/
│   ├── api/              # Interface: HTTP Handlers & Routes
│   ├── ws/               # Interface: WebSocket Signal Distribution
│   ├── core/             # Logic: Business Services & Workflows
│   ├── models/           # Logic: Domain Models & Rule Validation
│   ├── data/             # Data: SQL Repositories (sqlc)
│   ├── cache/            # Data: Redis & In-memory state
│   ├── workers/          # Operation: Background Jobs & CRONs
│   └── audit/            # Operation: Observability & Logging
└── Docs/                 # Comprehensive Technical Specification
```

---

## 🚀 Getting Started

### Prerequisites
*   [Go](https://go.dev/doc/install) (latest stable)
*   [Node.js](https://nodejs.org/) (v20+)
*   [PostgreSQL](https://www.postgresql.org/) & [Redis](https://redis.io/)

### Quick Start
1.  **Clone the repository**
    ```bash
    git clone https://github.com/your-username/encanto.git
    cd encanto
    ```

2.  **Setup Backend**
    ```bash
    cd backend
    cp .env.example .env # Configure your DB and Redis
    go mod download
    go run main.go
    ```

3.  **Setup Frontend**
    ```bash
    cd frontend
    npm install
    npm run dev
    ```

---

## 🧪 Testing

Encanto prioritizes quality through automated testing:
*   **E2E Tests**: `npx playwright test`
*   **Unit Tests**: `go test ./...`

---

## 📄 License & Status

Encanto is currently in its core development phase. For detailed implementation status, see [Milestone.md](Docs/Milestone.md).

---

Built with ❤️ by the Encanto Team.
