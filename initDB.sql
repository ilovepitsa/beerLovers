DROP TABLE IF EXISTS part_in_event;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS membership_fee;
DROP TABLE IF EXISTS donations;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS member;
DROP TABLE IF EXISTS wallet;
DROP TABLE IF EXISTS beer;
DROP TABLE IF EXISTS beer_type;


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
    email varchar(255) DEFAULT NULL,
    wallet_id integer references wallet(id),
    PRIMARY KEY(id)

);

CREATE TABLE beer_type(
    id serial,
    beer_type varchar(255) NOT NULL,
    description text DEFAULT NULL,
    PRIMARY KEY(id)
);

CREATE TABLE beer(
    id serial,
    name varchar(255) NOT NULL,
    producer varchar(255) NOT NULL,
    beer_type integer references beer_type(id),
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
    id serial,
    member_id integer references member(id),
    event_id integer references events(id),
    PRIMARY KEY(id)
);

CREATE TABLE sessions(
    id serial,
    member_id integer references member(id),
    PRIMARY KEY(id)
);