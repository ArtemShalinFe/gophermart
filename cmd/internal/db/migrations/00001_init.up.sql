begin transaction;

create type order_status as enum ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

-- Пользователи
create table users(
    id uuid default gen_random_uuid(),
    login varchar(200) unique not null,
    pass varchar(64) not null,
    
    primary key (id)
);

-- Заказы
create table orders(
    id uuid default gen_random_uuid(),
    uploaded timestamp with time zone not null,
    number numeric unique not null,
    userid uuid not null,
    sum double precision not null,
    status order_status not null,
    
    primary key (id),
    foreign key (userid) references users (id)
);

-- История списания баланса пользователя
create table withdrawals(
    seq bigserial,
    date timestamp with time zone not null,
    ordernumber numeric not null,
    userid uuid not null,
    sum double precision not null,
    
    primary key (seq),
    foreign key (userid) references users (id)
);

-- Для хранения текущего баланса, чтобы не считать по balances
create table currentbalances(
    seq bigserial,
    userid uuid unique not null,
    sum double precision not null,
    
    primary key (seq),
    foreign key (userid) references users (id)
);

commit;