-- name: ListVisibleChats :many
SELECT
  c.id,
  c.name,
  c.phone_number,
  c.status,
  c.last_message_preview,
  c.last_message_at,
  c.instance_id,
  wi.name AS instance_name,
  COALESCE(cus.is_hidden, false) AS is_hidden,
  COALESCE(cus.is_pinned, false) AS is_pinned
FROM contacts c
LEFT JOIN whatsapp_instances wi ON wi.id = c.instance_id
LEFT JOIN contact_user_states cus
  ON cus.contact_id = c.id
 AND cus.user_id = @viewer_user_id::uuid
WHERE c.organization_id = @organization_id::uuid
  AND c.deleted_at IS NULL
  AND (
    @scope_mode::text = 'all_contacts'
    OR (
      @scope_mode::text = 'instances_only'
      AND c.instance_id IS NOT NULL
      AND c.instance_id::text IN (
        SELECT jsonb_array_elements_text(@allowed_instance_ids::jsonb)
      )
    )
    OR (
      @scope_mode::text = 'allowed_numbers_only'
      AND c.phone_number IN (
        SELECT jsonb_array_elements_text(@allowed_phone_numbers::jsonb)
      )
    )
    OR (
      @scope_mode::text = 'instances_plus_allowed_numbers'
      AND (
        (
          c.instance_id IS NOT NULL
          AND c.instance_id::text IN (
            SELECT jsonb_array_elements_text(@allowed_instance_ids::jsonb)
          )
        )
        OR c.phone_number IN (
          SELECT jsonb_array_elements_text(@allowed_phone_numbers::jsonb)
        )
      )
    )
  )
  AND (@include_pending::bool OR c.status <> 'pending')
  AND (
    @search::text = ''
    OR COALESCE(c.name, '') ILIKE '%' || @search || '%'
    OR c.phone_number ILIKE '%' || @search || '%'
  )
ORDER BY COALESCE(cus.is_pinned, false) DESC, COALESCE(c.last_message_at, c.created_at) DESC;

-- name: GetVisibleChatByID :one
SELECT
  c.id,
  c.name,
  c.phone_number,
  c.status,
  c.last_message_preview,
  c.last_message_at,
  c.instance_id,
  wi.name AS instance_name,
  COALESCE(cus.is_hidden, false) AS is_hidden,
  COALESCE(cus.is_pinned, false) AS is_pinned
FROM contacts c
LEFT JOIN whatsapp_instances wi ON wi.id = c.instance_id
LEFT JOIN contact_user_states cus
  ON cus.contact_id = c.id
 AND cus.user_id = @viewer_user_id::uuid
WHERE c.organization_id = @organization_id::uuid
  AND c.id = @contact_id::uuid
  AND c.deleted_at IS NULL
  AND (
    @scope_mode::text = 'all_contacts'
    OR (
      @scope_mode::text = 'instances_only'
      AND c.instance_id IS NOT NULL
      AND c.instance_id::text IN (
        SELECT jsonb_array_elements_text(@allowed_instance_ids::jsonb)
      )
    )
    OR (
      @scope_mode::text = 'allowed_numbers_only'
      AND c.phone_number IN (
        SELECT jsonb_array_elements_text(@allowed_phone_numbers::jsonb)
      )
    )
    OR (
      @scope_mode::text = 'instances_plus_allowed_numbers'
      AND (
        (
          c.instance_id IS NOT NULL
          AND c.instance_id::text IN (
            SELECT jsonb_array_elements_text(@allowed_instance_ids::jsonb)
          )
        )
        OR c.phone_number IN (
          SELECT jsonb_array_elements_text(@allowed_phone_numbers::jsonb)
        )
      )
    )
  );

-- name: ListMessagesByContact :many
SELECT id, contact_id, direction, type, body, status, created_at, sent_by_user_id
FROM messages
WHERE organization_id = @organization_id::uuid
  AND contact_id = @contact_id::uuid
ORDER BY created_at ASC;

-- name: ListNotesByContact :many
SELECT
  n.id,
  n.contact_id,
  n.author_user_id,
  n.body,
  n.created_at,
  u.full_name AS author_name
FROM conversation_notes n
JOIN users u ON u.id = n.author_user_id
WHERE n.organization_id = @organization_id::uuid
  AND n.contact_id = @contact_id::uuid
ORDER BY n.created_at ASC;

