-- Creation of transaction table
--  PRIMARY KEY(date, type, symbol, account )
CREATE TABLE IF NOT EXISTS transactions (
    id  NUMERIC,
    date TIMESTAMP,
    type varchar(50),
    security varchar(255),
    security_payee varchar(255),
    symbol varchar(10),
    description varchar(255),
    shares NUMERIC,
    investment_amount NUMERIC,
    amount NUMERIC,
    account varchar(255),
    PRIMARY KEY(id)
);

CREATE TABLE IF NOT EXISTS all_transactions (
    id  NUMERIC,
    date TIMESTAMP,
    type varchar(50),
    security varchar(255),
    security_payee varchar(255),
    symbol varchar(10),
    description varchar(255),
    shares NUMERIC,
    investment_amount NUMERIC,
    amount NUMERIC,
    account varchar(255),
    PRIMARY KEY(id)
);

CREATE TABLE IF NOT EXISTS fund_history (
    symbol varchar(10),
    source varchar(50),
    date TIMESTAMP,
    open NUMERIC,
    high NUMERIC,
    low NUMERIC,
    close NUMERIC,
    adj_close NUMERIC,
    volume NUMERIC,
    PRIMARY KEY(symbol,date)
);

CREATE TABLE IF NOT EXISTS test_history (
    symbol varchar(10),
    source varchar(50),
    date TIMESTAMP,
    open NUMERIC,
    high NUMERIC,
    low NUMERIC,
    close NUMERIC,
    adj_close NUMERIC,
    volume NUMERIC,
    PRIMARY KEY(symbol,date)
    );

CREATE TABLE IF NOT EXISTS portfolio_value (
    date TIMESTAMP,
    name    VARCHAR(255),
    symbol VARCHAR(10),
    type VARCHAR(25),
    quote NUMERIC,
    pricedaychange NUMERIC,
    pricedaychangepct NUMERIC,
    shares NUMERIC,
    costbasis NUMERIC,
    marketvalue NUMERIC,
    averagecostpershare NUMERIC,
    gainloss12month NUMERIC,
    gainloss NUMERIC,
    gaillosspct NUMERIC,
    PRIMARY KEY(symbol,date)
);

CREATE TABLE IF NOT EXISTS dividends (
    ticker VARCHAR(25),
    cash_amount NUMERIC,
    declaration_date TIMESTAMP,
    dividend_type VARCHAR(25),
    ex_dividend_date TIMESTAMP,
    frequency NUMERIC,
    pay_date TIMESTAMP,
    record_date TIMESTAMP,
    PRIMARY KEY(ticker,declaration_date)
);

CREATE TABLE IF NOT EXISTS lookups (
    security VARCHAR(255),
    symbol VARCHAR(25),
    PRIMARY KEY(security)
);

CREATE TABLE IF NOT EXISTS dividend_history (
    symbol VARCHAR(25),
    year NUMERIC,
    month NUMERIC,
    amount NUMERIC,
    PRIMARY KEY(symbol,year,month)
);

-- CREATE INDEX IF NOT EXISTS ON dividend_history  USING (year,month);
-- CREATE INDEX IF NOT EXISTS year_mo_index ON dividend_history(year,month);
