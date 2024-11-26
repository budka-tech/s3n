create table bucket
(
    id          smallserial  not null,
    bucket_name varchar(255) not null,
    primary key (id),
    unique (bucket_name)
);

create table image
(
    id        uuid default gen_random_uuid() not null,
    bucket_id smallint                       not null,
    primary key (id),
    foreign key (bucket_id) references bucket
        on delete restrict
);

