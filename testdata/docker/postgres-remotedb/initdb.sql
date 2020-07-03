CREATE TYPE public.enum_type AS ENUM ('enum_one', 'enum_two', 'enum_three');
CREATE TABLE public.foo (
    id int,
    name text,
    enum_value public.enum_type
);
