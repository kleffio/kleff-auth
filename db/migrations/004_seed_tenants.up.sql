INSERT INTO tenants (slug, name)
VALUES ('kleff', 'Kleff Hosting'),
       ('portfolio', 'Isaac Portfolio')
ON CONFLICT (slug) DO NOTHING;
