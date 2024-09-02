DROP INDEX IF EXISTS idx_part_in_event_mid_eid;
DROP TABLE IF EXISTS part_in_event;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS membership_fee;
DROP TABLE IF EXISTS donations;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS member;
DROP TABLE IF EXISTS wallet;
DROP TABLE IF EXISTS beer;
DROP TABLE IF EXISTS beer_type;
DROP TYPE  IF EXISTS user_level;




CREATE TYPE user_level as enum('admin', 'user');


CREATE TABLE wallet(
    id serial,
    balance numeric(15,2) DEFAULT 0.00,
    PRIMARY KEY(id)
);

CREATE TABLE member(
    id serial,
    fio varchar(255) NOT NULL,
    entry_date date NOT NULL,
    address varchar(255) DEFAULT NULL,
    phone_number varchar(255) DEFAULT NULL,
    email varchar(255) NOT NULL,
    password bytea NOT NULL,
    wallet_id integer references wallet(id),
    level user_level NOT NULL,
    PRIMARY KEY(id)

);

CREATE TABLE beer_type(
    id serial,
    type_name varchar(255) NOT NULL,
    PRIMARY KEY(id)
);

CREATE TABLE beer(
    id serial,
    name varchar(255) NOT NULL,
    producer varchar(255) NOT NULL,
    beer_type integer references beer_type(id),
    photo_url text,
    PRIMARY KEY(id)
);

CREATE TABLE events(
    id serial,
    name varchar(255) NOT NULL,
    date date NOT NULL,
    location text DEFAULT NULL,
    description text DEFAULT NULL,
    PRIMARY KEY(id)
);

CREATE TABLE membership_fee(
    id serial,
    member_id integer references member(id),
    value numeric(15,2) NOT NULL,
    date date NOT NULL,
    PRIMARY KEY(id)
);

CREATE TABLE donations(
    id serial,
    member_id integer references member(id),
    value numeric(15,2) NOT NULL,
    date date NOT NULL,
    PRIMARY KEY(id)
);

CREATE TABLE part_in_event(
    member_id integer references member(id),
    event_id integer references events(id)
);
CREATE UNIQUE INDEX idx_part_in_event_mid_eid on part_in_event (member_id, event_id);

CREATE TABLE sessions(
    id varchar(32),
    member_id integer references member(id),
    PRIMARY KEY(id)
);

INSERT INTO wallet (balance) VALUES(0);
INSERT INTO member (id, fio, entry_date, email, password, wallet_id, level)
VALUES (DEFAULT, 'admin', DATE'2001-09-29', 'admin@admin.ru', decode('5254665a5541636b94b63ab83bb6a172b7ad541ad1cbdcdb2e84e1b7ce6d2c521135852b9a7ffe9c','hex'), 1, 'admin');


insert into events (name, date, location, description) values ('Приветственная вечеринка', '2024-08-01', 'Дом', 'Приветственная вечеринка для новичков');  
insert into events (name, date, location, description) values ('ВВВ', '2002-04-30', 'Владивосток', 'Вечеринка во Владивостоке');  
insert into events (name, date, location, description) values ('МММ', '2024-09-29', 'Москва', 'Московская Мужская Мочиловка');  
insert into beer_type (type_name) values ('Пейл-эль'),
('Пшеничный эль'),
('Бельгийский эль'),
('Кислый эль'),
('Бурый эль'),
('Портер'),
('Стаут'),
('Светлый лагер'),
('Темный лагер'),
('Бок'),
('Янтарное пиво'),
('Специальный сорт');
