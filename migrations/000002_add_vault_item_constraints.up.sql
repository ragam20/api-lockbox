ALTER TABLE vault_items
DROP CONSTRAINT IF EXISTS chk_vault_items_item_type;

ALTER TABLE vault_items
ADD CONSTRAINT chk_vault_items_item_type
CHECK (
    item_type IN ('password', 'note', 'apiKey')
);
