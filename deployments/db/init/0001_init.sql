CREATE USER gopher
    PASSWORD 'gopher';

CREATE DATABASE gophermart
    OWNER 'gopher'
    ENCODING 'UTF8'
    LC_COLLATE = 'en_US.utf8'
    LC_CTYPE = 'en_US.utf8';

CREATE DATABASE accrual
    OWNER 'gopher'
    ENCODING 'UTF8'
    LC_COLLATE = 'en_US.utf8'
    LC_CTYPE = 'en_US.utf8';