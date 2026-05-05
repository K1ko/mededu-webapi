const mongoHost = process.env.MEDEDU_API_MONGODB_HOST
const mongoPort = process.env.MEDEDU_API_MONGODB_PORT

const mongoUser = process.env.MEDEDU_API_MONGODB_USERNAME
const mongoPassword = process.env.MEDEDU_API_MONGODB_PASSWORD

const database = process.env.MEDEDU_API_MONGODB_DATABASE
const collection = process.env.MEDEDU_API_MONGODB_COLLECTION

const retrySeconds = parseInt(process.env.RETRY_CONNECTION_SECONDS || "5") || 5;

let connection;
while (true) {
    try {
        const auth = mongoUser ? `${encodeURIComponent(mongoUser)}:${encodeURIComponent(mongoPassword)}@` : "";
        connection = Mongo(`mongodb://${auth}${mongoHost}:${mongoPort}/admin`);
        break;
    } catch (exception) {
        print(`Cannot connect to mongoDB: ${exception}`);
        print(`Will retry after ${retrySeconds} seconds`);
        sleep(retrySeconds * 1000);
    }
}

const databases = connection.getDBNames();
if (databases.includes(database)) {
    const dbInstance = connection.getDB(database);
    const collections = dbInstance.getCollectionNames();
    if (collections.includes(collection)) {
        print(`Collection '${collection}' already exists in database '${database}'`);
        process.exit(0);
    }
}

const db = connection.getDB(database);
db.createCollection(collection);
db[collection].createIndex({ "id": 1 }, { unique: true });
db[collection].createIndex({ "department": 1 });
db[collection].createIndex({ "startAt": 1 });

const result = db[collection].insertMany([
    {
        "id": "urgent-safety-2026-05",
        "title": "BOZP pre urgentny prijem",
        "type": "mandatory",
        "department": "Urgent",
        "startAt": new Date("2026-05-20T08:00:00Z"),
        "capacity": 20,
        "lecturer": "Mgr. Jana Novakova",
        "location": "Skoliaca miestnost A",
        "onlineLink": "",
        "description": "Interne skolenie pre personal urgentneho prijmu.",
        "requirements": "Zamestnanecky preukaz",
        "status": "planned",
        "occupied": 16,
        "waitlisted": 0
    },
    {
        "id": "icu-infection-2026-06",
        "title": "Prevencia infekcii na JIS",
        "type": "department",
        "department": "JIS",
        "startAt": new Date("2026-06-02T12:30:00Z"),
        "capacity": 12,
        "lecturer": "MUDr. Eva Hruba",
        "location": "",
        "onlineLink": "https://teams.example/mededu/icu-infection",
        "description": "Prakticke postupy prevencie nozokomialnych infekcii.",
        "requirements": "Notebook alebo tablet",
        "status": "planned",
        "occupied": 12,
        "waitlisted": 3
    }
]);

if (result.writeError) {
    console.error(result);
    print(`Error when writing the data: ${result.errmsg}`);
}

process.exit(0);
