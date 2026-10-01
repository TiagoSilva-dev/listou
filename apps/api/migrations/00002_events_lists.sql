-- +goose Up
CREATE TABLE events (
    id                   UUID PRIMARY KEY,
    owner_id             UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type                 TEXT NOT NULL CHECK (type IN ('BABY_SHOWER', 'WEDDING', 'HOUSEWARMING', 'BIRTHDAY', 'GRADUATION', 'TRAVEL', 'CHRISTMAS', 'WISHLIST', 'CUSTOM')),
    title                TEXT NOT NULL CHECK (char_length(title) BETWEEN 1 AND 120),
    slug                 TEXT NOT NULL CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND char_length(slug) BETWEEN 3 AND 60),
    description          TEXT,
    host_names           TEXT,
    event_date           DATE,
    location             TEXT,
    cover_image_url      TEXT,
    avatar_url           TEXT,
    theme                TEXT NOT NULL DEFAULT 'blush',
    visibility           TEXT NOT NULL DEFAULT 'UNLISTED' CHECK (visibility IN ('PUBLIC', 'UNLISTED', 'PRIVATE')),
    status               TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'PUBLISHED', 'ARCHIVED')),
    surprise_mode        BOOLEAN NOT NULL DEFAULT false,
    show_reserver_names  BOOLEAN NOT NULL DEFAULT false,
    published_at         TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at           TIMESTAMPTZ
);
-- Slugs stay reserved even for soft-deleted events so old links never point
-- at someone else's list.
CREATE UNIQUE INDEX events_slug_key ON events (slug);
CREATE INDEX events_owner_idx ON events (owner_id, created_at DESC) WHERE deleted_at IS NULL;

-- Co-owners and future collaborators (GiftListMember in the domain model).
CREATE TABLE event_members (
    id         UUID PRIMARY KEY,
    event_id   UUID NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role       TEXT NOT NULL CHECK (role IN ('CO_OWNER', 'VIEWER')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT event_members_event_user_key UNIQUE (event_id, user_id)
);
CREATE INDEX event_members_user_idx ON event_members (user_id);

CREATE TABLE gift_lists (
    id                        UUID PRIMARY KEY,
    event_id                  UUID NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    title                     TEXT NOT NULL,
    description               TEXT,
    is_primary                BOOLEAN NOT NULL DEFAULT true,
    allow_reservations        BOOLEAN NOT NULL DEFAULT true,
    allow_group_contributions BOOLEAN NOT NULL DEFAULT false,
    status                    TEXT NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'ARCHIVED')),
    created_at                TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX gift_lists_event_idx ON gift_lists (event_id);
CREATE UNIQUE INDEX gift_lists_one_primary_per_event ON gift_lists (event_id) WHERE is_primary;

CREATE TABLE categories (
    id           UUID PRIMARY KEY,
    gift_list_id UUID NOT NULL REFERENCES gift_lists (id) ON DELETE CASCADE,
    name         TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 60),
    emoji        TEXT,
    position     INT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT categories_list_name_key UNIQUE (gift_list_id, name)
);

-- +goose Down
DROP TABLE categories;
DROP TABLE gift_lists;
DROP TABLE event_members;
DROP TABLE events;
