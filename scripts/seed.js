const { Client } = require('pg');
const dotenv = require('dotenv');
const { randomUUID } = require('crypto');
const path = require('path');

// Load environment variables from the root .env file
dotenv.config({ path: path.join(__dirname, '../.env') });

const config = {
    host: process.env.AEGIS_DB_HOST || '127.0.0.1',
    port: parseInt(process.env.AEGIS_DB_PORT || '5433'),
    user: process.env.AEGIS_DB_USER || 'aegis',
    password: process.env.AEGIS_DB_PASSWORD || 'aegis_secret',
    database: process.env.AEGIS_DB_NAME || 'aegis',
};

const client = new Client(config);

async function seed(count = 10) {
    try {
        await client.connect();
        console.log('Connected to database successfully');

        for (let i = 0; i < count; i++) {
            const userId = randomUUID();
            const platformId = `PLAT-${Math.floor(Math.random() * 100000)}`;
            const pspId = `PSP-${Math.floor(Math.random() * 100000)}`;
            const name = `Test User ${i + 1}`;
            const email = `user${i + 1}@example.com`;

            // Insert User
            await client.query(
                `INSERT INTO users (id, platform_id, psp_id, name, email, created_at, updated_at) 
                 VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`,
                [userId, platformId, pspId, name, email]
            );

            // Insert Wallet
            const walletId = randomUUID();
            await client.query(
                `INSERT INTO wallets (id, user_id, type, balance, locked_balance, currency, created_at, updated_at) 
                 VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`,
                [walletId, userId, 'settlement', 0, 0, 'GHS']
            );

            process.stdout.write(`\rSeeded ${i + 1}/${count} users and wallets...`);
        }

        console.log('\nSeeding completed successfully!');
    } catch (err) {
        console.error('\nSeeding failed:', err.message);
    } finally {
        await client.end();
    }
}

// Get count from command line argument or default to 10
const countArg = process.argv[2] ? parseInt(process.argv[2]) : 10;
seed(countArg);
