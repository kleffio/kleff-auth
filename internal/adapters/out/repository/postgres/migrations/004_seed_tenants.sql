-- +goose Up
INSERT INTO tenants (slug, name)
VALUES ('kleff', 'Kleff Hosting'),
       ('portfolio', 'Isaac Portfolio')
ON CONFLICT (slug) DO NOTHING;

-- +goose Down
DELETE FROM tenants WHERE slug IN ('kleff','portfolio');
