-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users.roles (
    name TEXT NOT NULL,
    CONSTRAINT pk_roles PRIMARY KEY (name)
);

CREATE TABLE IF NOT EXISTS users.permissions (
    code TEXT NOT NULL,
    CONSTRAINT pk_permissions PRIMARY KEY (code)
);

CREATE TABLE IF NOT EXISTS users.role_permissions (
    permission_code TEXT NOT NULL,
    role_name       TEXT NOT NULL,
    CONSTRAINT pk_role_permissions PRIMARY KEY (permission_code, role_name),
    CONSTRAINT fk_role_permissions_permission FOREIGN KEY (permission_code) REFERENCES users.permissions (code) ON DELETE CASCADE,
    CONSTRAINT fk_role_permissions_role       FOREIGN KEY (role_name)       REFERENCES users.roles (name)       ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS users.user_roles (
    user_id   UUID NOT NULL,
    role_name TEXT NOT NULL,
    CONSTRAINT pk_user_roles PRIMARY KEY (user_id, role_name),
    CONSTRAINT fk_user_roles_user FOREIGN KEY (user_id)   REFERENCES users.users (id)   ON DELETE CASCADE,
    CONSTRAINT fk_user_roles_role FOREIGN KEY (role_name) REFERENCES users.roles (name) ON DELETE CASCADE
);

-- seed roles
INSERT INTO users.roles (name) VALUES ('Administrator'), ('Member') ON CONFLICT DO NOTHING;

-- seed permissions
INSERT INTO users.permissions (code) VALUES
    ('users:read'),
    ('users:update'),
    ('events:read'),
    ('events:search'),
    ('events:update'),
    ('ticket-types:read'),
    ('ticket-types:update'),
    ('categories:read'),
    ('categories:update'),
    ('carts:read'),
    ('carts:add'),
    ('carts:remove'),
    ('orders:read'),
    ('orders:create'),
    ('tickets:read'),
    ('tickets:check-in'),
    ('event-statistics:read')
ON CONFLICT DO NOTHING;

-- seed role_permissions — Member
INSERT INTO users.role_permissions (permission_code, role_name) VALUES
    ('users:read',        'Member'),
    ('users:update',      'Member'),
    ('events:search',     'Member'),
    ('ticket-types:read', 'Member'),
    ('carts:read',        'Member'),
    ('carts:add',         'Member'),
    ('carts:remove',      'Member'),
    ('orders:read',       'Member'),
    ('orders:create',     'Member'),
    ('tickets:read',      'Member'),
    ('tickets:check-in',  'Member')
ON CONFLICT DO NOTHING;

-- seed role_permissions — Administrator (everything)
INSERT INTO users.role_permissions (permission_code, role_name) VALUES
    ('users:read',             'Administrator'),
    ('users:update',           'Administrator'),
    ('events:read',            'Administrator'),
    ('events:search',          'Administrator'),
    ('events:update',          'Administrator'),
    ('ticket-types:read',      'Administrator'),
    ('ticket-types:update',    'Administrator'),
    ('categories:read',        'Administrator'),
    ('categories:update',      'Administrator'),
    ('carts:read',             'Administrator'),
    ('carts:add',              'Administrator'),
    ('carts:remove',           'Administrator'),
    ('orders:read',            'Administrator'),
    ('orders:create',          'Administrator'),
    ('tickets:read',           'Administrator'),
    ('tickets:check-in',       'Administrator'),
    ('event-statistics:read',  'Administrator')
ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users.user_roles;
DROP TABLE IF EXISTS users.role_permissions;
DROP TABLE IF EXISTS users.permissions;
DROP TABLE IF EXISTS users.roles;
-- +goose StatementEnd
