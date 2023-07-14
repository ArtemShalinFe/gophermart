BEGIN TRANSACTION;

-- Пользователи
CREATE TABLE Users(
    id uuid default gen_random_uuid(),
    login varchar(200) unique not null,
    pass varchar(64) not null,
    
    primary key (id)
);

-- Заказы
CREATE TABLE Orders(
    id uuid default gen_random_uuid(),
    uploaded timestamp with time zone,
    number numeric unique not null,
    userId uuid not NULL,
    sum double precision NOT null,
    status VARCHAR(16) not null,
    
    primary key (id),
    foreign key (userId) references Users (id)
);

-- История начисления и списания баланса пользователя
CREATE TABLE Balances(
    date timestamp with time zone,
	userId uuid not null,
	sum integer NOT null,
	
	foreign key (userId) references Users (id)
);

-- Для хранения текущего баланса, чтобы не считать по Balances
CREATE TABLE CurrentBalances(
	userId uuid not null,
	sum integer not null,
	
	foreign key (userId) references Users (id)
);

COMMIT;