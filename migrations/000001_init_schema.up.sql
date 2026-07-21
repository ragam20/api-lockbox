--
-- PostgreSQL database dump
--

-- Dumped from database version 16.13
-- Dumped by pg_dump version 16.14 (Ubuntu 16.14-0ubuntu0.24.04.1)

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
-- Name: argon_config; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.argon_config (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    "time" bigint DEFAULT 3 NOT NULL,
    memory bigint DEFAULT 65536 NOT NULL,
    parallelism bigint DEFAULT 2 NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT chk_argon_config_memory CHECK (((memory >= 65536) AND (memory <= 131072))),
    CONSTRAINT chk_argon_config_parallelism CHECK (((parallelism >= 1) AND (parallelism <= 7))),
    CONSTRAINT chk_argon_config_time CHECK ((("time" >= 1) AND ("time" <= 8)))
);


--
-- Name: argon_config_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.argon_config_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: argon_config_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.argon_config_id_seq OWNED BY public.argon_config.id;


--
-- Name: user_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_sessions (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    device character varying(255),
    refresh_token_hash text NOT NULL,
    user_agent text,
    ip_address text,
    last_used_at timestamp with time zone,
    revoke_at timestamp with time zone,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: user_sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_sessions_id_seq OWNED BY public.user_sessions.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    name character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    email_verified boolean,
    public_img_id text,
    img_url text,
    master_key_salt text NOT NULL,
    auth_verifier text NOT NULL,
    auth_verifier_salt text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: vault_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.vault_items (
    id bigint NOT NULL,
    vault_id bigint NOT NULL,
    title text,
    item_type character varying(20) NOT NULL,
    iv_ciphertext text NOT NULL,
    ciphertext text NOT NULL,
    iv_protection_key text NOT NULL,
    protection_key text NOT NULL,
    favorite_at timestamp with time zone,
    archived_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT chk_vault_items_item_type CHECK (((item_type)::text = ANY ((ARRAY['password'::character varying, 'note'::character varying])::text[])))
);


--
-- Name: vault_items_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.vault_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: vault_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.vault_items_id_seq OWNED BY public.vault_items.id;


--
-- Name: vaults; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.vaults (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    timeout integer DEFAULT 10,
    timeout_type character varying(20),
    vault_key_salt text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT chk_vaults_timeout_type CHECK (((timeout_type)::text = ANY ((ARRAY['lock'::character varying, 'logout'::character varying])::text[])))
);


--
-- Name: vaults_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.vaults_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: vaults_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.vaults_id_seq OWNED BY public.vaults.id;


--
-- Name: argon_config id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.argon_config ALTER COLUMN id SET DEFAULT nextval('public.argon_config_id_seq'::regclass);


--
-- Name: user_sessions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_sessions ALTER COLUMN id SET DEFAULT nextval('public.user_sessions_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: vault_items id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.vault_items ALTER COLUMN id SET DEFAULT nextval('public.vault_items_id_seq'::regclass);


--
-- Name: vaults id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.vaults ALTER COLUMN id SET DEFAULT nextval('public.vaults_id_seq'::regclass);


--
-- Name: argon_config argon_config_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.argon_config
    ADD CONSTRAINT argon_config_pkey PRIMARY KEY (id);


--
-- Name: user_sessions user_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: vault_items vault_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.vault_items
    ADD CONSTRAINT vault_items_pkey PRIMARY KEY (id);


--
-- Name: vaults vaults_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.vaults
    ADD CONSTRAINT vaults_pkey PRIMARY KEY (id);


--
-- Name: idx_argon_config_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_argon_config_user_id ON public.argon_config USING btree (user_id);


--
-- Name: idx_user_sessions_expires_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_sessions_expires_at ON public.user_sessions USING btree (expires_at);


--
-- Name: idx_user_sessions_refresh_token_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_user_sessions_refresh_token_hash ON public.user_sessions USING btree (refresh_token_hash);


--
-- Name: idx_user_sessions_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_sessions_user_id ON public.user_sessions USING btree (user_id);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: idx_vault_items_vault_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_vault_items_vault_id ON public.vault_items USING btree (vault_id);


--
-- Name: idx_vaults_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_vaults_user_id ON public.vaults USING btree (user_id);


--
-- Name: argon_config fk_argon_config_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.argon_config
    ADD CONSTRAINT fk_argon_config_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_sessions fk_user_sessions_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT fk_user_sessions_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: vault_items fk_vault_items_vault; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.vault_items
    ADD CONSTRAINT fk_vault_items_vault FOREIGN KEY (vault_id) REFERENCES public.vaults(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: vaults fk_vaults_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.vaults
    ADD CONSTRAINT fk_vaults_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--
