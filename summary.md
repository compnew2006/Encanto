# Summary

**Task:** صمّم نموذج صلاحيات قائم على الأفعال لا على أسماء الأدوار فقط، بمخرجات قابلة للربط المباشر مع الأزرار والشاشات والطلبات، مع تغطية حالات المشاهدة بدون كتابة.

**Approach & Key Decisions:**
- Created `Docs/15_permissions_action_model.md` to formally document the Action-Based Permissions model.
- Analyzed existing UI surfaces from `10_screen_behavior_reference.md` and database structures from `03_database_schema.md` to ensure the mapping matches actual operational requirements.
- Standardized the format `[Resource].[Action]` (e.g., `chats.view`, `messages.send`).
- Addressed the specific requirement "حالات تسمح بالمشاهدة وتمنع الكتابة" by ensuring actions are de-coupled (granting `view` while withholding `send/create/edit`).
- Provided implementation examples for Frontend UI (e.g., UI Conditional rendering buttons) and Backend Middleware (Endpoint method checks).

**Files Modified/Created/Deleted:**
- [NEW] `Docs/15_permissions_action_model.md` (Created action-based permission schema).
- [MODIFIED] `summary.md` (Overwritten with updated task output).

**Tests Added:**
- N/A (Documentation and Architecture design phase).

**Results & Verification:**
- An action-based capabilities model is ready, fulfilling requirements to separate roles from hardcoded UI logic and linking directly to API methods and interface actions.
