# Summary

**Task:** وضع طبقة رؤية واضحة (Visibility Scope Layer) تحدد سياسات استعراض البيانات (من يرى كل العناصر، المحددة بنطاق، أو أرقام مختارة) لتعميمها بقوة على القوائم، التفاصيل، البحث، والتصدير.

**Approach & Key Decisions:**
- Created `Docs/16_visibility_scope_layer.md` to define unified database-enforced filtering for `user_contact_visibility_rules`.
- Defined four main scopes: `all_contacts` (الرؤية الشاملة), `instances_only` (حسب الحسابات المرتبطة), `allowed_numbers_only` (أرقام مستثناة فردياً), and `instances_plus_allowed_numbers` (مدمجة).
- Documented Enforcement Levels ensuring that a uniform SQL middleware dictates what comes out from `Lists/Inbox`, `Detail Queries (403/404 handling)`, `Deep Search`, and `CSV Exports`.
- Included the concept of hierarchy resolution (Inherit from Roles vs. Direct Override) and Data Masking (`can_view_unmasked_phone = false`) which masks numbers equally across the UI and CSV chunks.

**Files Modified/Created/Deleted:**
- [NEW] `Docs/16_visibility_scope_layer.md` (Created visibility scoping blueprint).
- [MODIFIED] `summary.md` (Overwritten with updated task output).

**Tests Added:**
- N/A (Architecture and Security scoping phase).

**Results & Verification:**
- A robust, query-level data isolation architecture is structurally documented, fulfilling the exact restrictions mapping required for multi-tier role implementations.
