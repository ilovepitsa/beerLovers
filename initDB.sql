-- DROP INDEX IF EXISTS idx_part_in_event_mid_eid;
DROP TABLE IF EXISTS part_in_event;
DROP TABLE IF EXISTS favorite_beer;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS review;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS member;
DROP TABLE IF EXISTS beer;
DROP TABLE IF EXISTS beer_type;
DROP TYPE  IF EXISTS user_level;




CREATE TYPE user_level as enum('admin', 'user');

CREATE TABLE member(
    id serial,
    fio varchar(255) NOT NULL,
    entry_date date NOT NULL,
    email varchar(255) NOT NULL,
    password bytea NOT NULL,
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

CREATE TABLE favorite_beer(
    member_id integer references member(id),
    beer_id integer references beer(id),
    PRIMARY KEY(member_id, beer_id)
);

CREATE TABLE events(
    id serial,
    name varchar(255) NOT NULL,
    date date NOT NULL,
    responsible integer references member(id),
    location text DEFAULT NULL,
    description text DEFAULT NULL,
    PRIMARY KEY(id)
);

CREATE TABLE part_in_event(
    member_id integer references member(id),
    event_id integer references events(id),
    PRIMARY KEY(member_id, event_id)
);

CREATE TABLE review(
    id serial,
    event_id integer references events(id),
    member_id integer references member(id),
    text text NOT NULL,
    photo_url text,
    PRIMARY KEY(id)
);

CREATE TABLE sessions(
    id varchar(32),
    member_id integer references member(id),
    PRIMARY KEY(id)
);

INSERT INTO member (fio, entry_date, email, password, level)
VALUES ('Никита Щугорев', DATE'2001-09-29', 'admin@admin.ru', decode('5254665a5541636b94b63ab83bb6a172b7ad541ad1cbdcdb2e84e1b7ce6d2c521135852b9a7ffe9c','hex'), 'admin'),
('Дарья Логинова', DATE'2001-04-30', 'dasha@admin.ru', decode('5254665a5541636b94b63ab83bb6a172b7ad541ad1cbdcdb2e84e1b7ce6d2c521135852b9a7ffe9c','hex'), 'user');


insert into events (name, date, location, description, responsible) values ('Приветственная вечеринка', '2024-08-01', 'Дом', 'Приветственная вечеринка для новичков',1);  
insert into events (name, date, location, description, responsible) values ('ВВВ', '2002-04-30', 'Владивосток', 'Вечеринка во Владивостоке',1);  
insert into events (name, date, location, description, responsible) values ('МММ', '2024-09-29', 'Москва', 'Московская Мужская Мочиловка',1);  
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

insert into beer (name, producer, beer_type, photo_url) values ('Hoegaarden', 'Brouwerij Hoegaarden',2, '3422354ab003394e2ba8b0259e951dd8_res'),
('Балтика 9','Carlsberg Group',8, '0ea92de644f626689d231ae4dcba22d8_res' );