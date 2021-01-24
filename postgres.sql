--
-- PostgreSQL database dump
--

-- Dumped from database version 12.4 (Debian 12.4-1.pgdg100+1)
-- Dumped by pg_dump version 12.4 (Debian 12.4-1.pgdg100+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: asset_data; Type: TABLE; Schema: public; Owner: texas_real_foods
--

CREATE TABLE public.asset_data (
    business_id uuid NOT NULL,
    uri text,
    website_live boolean DEFAULT true,
    phone text[] DEFAULT '{}'::text[]
);


ALTER TABLE public.asset_data OWNER TO texas_real_foods;

--
-- Name: asset_metadata; Type: TABLE; Schema: public; Owner: texas_real_foods
--

CREATE TABLE public.asset_metadata (
    business_id uuid NOT NULL,
    business_name text NOT NULL,
    last_update timestamp without time zone DEFAULT now() NOT NULL,
    added timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.asset_metadata OWNER TO texas_real_foods;

--
-- Name: asset_updates; Type: TABLE; Schema: public; Owner: texas_real_foods
--

CREATE TABLE public.asset_updates (
    business_id uuid NOT NULL,
    fields_updated text[] NOT NULL,
    "timestamp" timestamp without time zone NOT NULL
);


ALTER TABLE public.asset_updates OWNER TO texas_real_foods;

--
-- Name: notifications; Type: TABLE; Schema: public; Owner: texas_real_foods
--

CREATE TABLE public.notifications (
    notification_id uuid NOT NULL,
    event_timestamp timestamp without time zone DEFAULT now() NOT NULL,
    notification json NOT NULL,
    read boolean DEFAULT false
);


ALTER TABLE public.notifications OWNER TO texas_real_foods;

--
-- Name: asset_updates asset_updates_pkey; Type: CONSTRAINT; Schema: public; Owner: texas_real_foods
--

ALTER TABLE ONLY public.asset_updates
    ADD CONSTRAINT asset_updates_pkey PRIMARY KEY (business_id);


--
-- Name: asset_data assets_data_pkey; Type: CONSTRAINT; Schema: public; Owner: texas_real_foods
--

ALTER TABLE ONLY public.asset_data
    ADD CONSTRAINT assets_data_pkey PRIMARY KEY (business_id);


--
-- Name: asset_metadata assets_meta_pkey; Type: CONSTRAINT; Schema: public; Owner: texas_real_foods
--

ALTER TABLE ONLY public.asset_metadata
    ADD CONSTRAINT assets_meta_pkey PRIMARY KEY (business_id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: texas_real_foods
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (notification_id);


--
-- PostgreSQL database dump complete
--

