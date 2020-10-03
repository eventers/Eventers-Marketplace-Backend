create table Users
(
    user_id int(21) auto_increment
        primary key,
    first_name varchar(45) null,
    last_name varchar(45) null,
    display_name varchar(100) null,
    email_address varchar(100) null,
    city varchar(45) null,
    state varchar(45) null,
    country varchar(45) null,
    address varchar(400) null,
    pincode varchar(20) null,
    fb_firebase_id varchar(100) null,
    fb_email varchar(100) null,
    fb_name varchar(45) null,
    fb_image_url varchar(100) null,
    g_firebase_id varchar(100) null,
    g_email varchar(100) null,
    g_name varchar(45) null,
    g_image_url varchar(100) null,
    phone_firebase_id varchar(100) null,
    phone_country_code varchar(5) null,
    phone_number varchar(15) null,
    provider varchar(15) null,
    created_date datetime default CURRENT_TIMESTAMP not null,
    updated_date datetime default CURRENT_TIMESTAMP null,
    profile_pic varchar(100) null,
    is_registered tinyint unsigned default 0 not null,
    is_active tinyint unsigned default 0 not null,
    is_force_stop tinyint unsigned default 0 null,
    a_firebase_id varchar(100) null,
    a_email varchar(100) null,
    a_name varchar(45) null,
    a_image_url varchar(100) null
);

create table Public_Event
(
    public_event_id int(21) auto_increment
        primary key,
    date_time datetime null,
    event_title varchar(1000) null,
    event_description varchar(2048) null,
    event_image varchar(1000) null,
    total_tickets int(100) null,
    ticket_price int(100) null,
    temp_account_address varchar(1024) null,
    temp_security_paraphrase varchar(1024) null
);

create table Event_Tickets
(
    event_ticket_id int(21) auto_increment
        primary key,
    business_user_id int(21) null,
    public_event_id int(21) null,
    asset_id varchar(100) null,
    current_holder_id int(21) null,
    status varchar(50) null,
    available_to_resell tinyint(1) default 0 null,
    price int null
);