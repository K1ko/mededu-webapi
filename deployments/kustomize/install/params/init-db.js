const mongoHost = process.env.MEDEDU_API_MONGODB_HOST
const mongoPort = process.env.MEDEDU_API_MONGODB_PORT

const mongoUser = process.env.MEDEDU_API_MONGODB_USERNAME
const mongoPassword = process.env.MEDEDU_API_MONGODB_PASSWORD

const database = process.env.MEDEDU_API_MONGODB_DATABASE
const collection = process.env.MEDEDU_API_MONGODB_COLLECTION

const retrySeconds = parseInt(process.env.RETRY_CONNECTION_SECONDS || "5") || 5;
const seedMode = (process.env.MEDEDU_API_MONGODB_SEED_MODE || "create-only").toLowerCase();

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
        if (seedMode === "reset") {
            print(`Dropping existing collection '${collection}' in database '${database}' before demo seed`);
            dbInstance[collection].drop();
        } else {
            print(`Collection '${collection}' already exists in database '${database}'`);
            process.exit(0);
        }
    }
}

const db = connection.getDB(database);
db.createCollection(collection);
db[collection].createIndex({ "id": 1 }, { unique: true });
db[collection].createIndex({ "department": 1 });
db[collection].createIndex({ "startAt": 1 });
db[collection].createIndex({ "registrations.employeeId": 1 });

function registration(id, employeeId, employeeName, department, registeredAt, extra = {}) {
    return {
        "id": id,
        "employeeId": employeeId,
        "employeeName": employeeName,
        "employeeEmail": extra.employeeEmail || `${employeeId.toLowerCase()}@hospital.example`,
        "department": department,
        "note": extra.note || "",
        "registeredAt": new Date(registeredAt)
    };
}

function training(input) {
    const registrations = (input.registrations || [])
        .sort((left, right) => left.registeredAt - right.registeredAt)
        .map((item, index) => ({
            ...item,
            "trainingId": input.id,
            "status": index < input.capacity ? "registered" : "waitlisted"
        }));

    return {
        "id": input.id,
        "title": input.title,
        "type": input.type,
        "department": input.department,
        "startAt": new Date(input.startAt),
        "capacity": input.capacity,
        "lecturer": input.lecturer,
        "location": input.location || "",
        "onlineLink": input.onlineLink || "",
        "description": input.description || "",
        "requirements": input.requirements || "",
        "status": input.status || "planned",
        "occupied": registrations.filter(item => item.status === "registered").length,
        "waitlisted": registrations.filter(item => item.status === "waitlisted").length,
        "registrations": registrations
    };
}

const result = db[collection].insertMany([
    training({
        "id": "urgent-safety-2026-05",
        "title": "BOZP pre urgentný príjem",
        "type": "mandatory",
        "department": "Urgent",
        "startAt": "2026-05-20T08:00:00Z",
        "capacity": 20,
        "lecturer": "Mgr. Jana Nováková",
        "location": "Školiaca miestnosť A",
        "description": "Povinné školenie bezpečnosti práce pre personál urgentného príjmu.",
        "requirements": "Zamestnanecký preukaz",
        "registrations": [
            registration("reg-urgent-001", "EMP-1042", "Bc. Peter Malina", "Urgent", "2026-05-01T09:15:00Z", { "note": "Uprednostňuje ranný termín." }),
            registration("reg-urgent-002", "EMP-1077", "Mgr. Lucia Križová", "Chirurgia", "2026-05-02T11:10:00Z"),
            registration("reg-urgent-003", "EMP-1091", "MUDr. Martin Havel", "Urgent", "2026-05-03T07:40:00Z"),
            registration("reg-urgent-004", "EMP-1150", "Bc. Nina Bartošová", "Urgent", "2026-05-04T14:25:00Z"),
            registration("reg-urgent-005", "EMP-1216", "Mgr. Adam Kováč", "Interné", "2026-05-05T08:05:00Z"),
            registration("reg-urgent-006", "EMP-1320", "MUDr. Petra Malíková", "JIS", "2026-05-06T12:45:00Z")
        ]
    }),
    training({
        "id": "icu-infection-2026-06",
        "title": "Prevencia infekcií na JIS",
        "type": "department",
        "department": "JIS",
        "startAt": "2026-06-02T12:30:00Z",
        "capacity": 2,
        "lecturer": "MUDr. Eva Hrubá",
        "onlineLink": "https://teams.example/mededu/icu-infection",
        "description": "Praktické postupy prevencie nozokomiálnych infekcií a práca s izolačným režimom.",
        "requirements": "Notebook alebo tablet",
        "registrations": [
            registration("reg-icu-001", "EMP-2018", "MUDr. Tomáš Benko", "JIS", "2026-05-05T07:30:00Z"),
            registration("reg-icu-002", "EMP-2034", "Mgr. Andrea Poláková", "JIS", "2026-05-06T10:20:00Z"),
            registration("reg-icu-003", "EMP-2071", "Bc. Katarína Švecová", "JIS", "2026-05-07T09:45:00Z", { "note": "Môže prísť aj na náhradný termín." }),
            registration("reg-icu-004", "EMP-2144", "MUDr. Jozef Varga", "Anestéziológia", "2026-05-08T13:05:00Z")
        ]
    }),
    training({
        "id": "pediatrics-onboarding-2026-06",
        "title": "Adaptačné školenie pre pediatriu",
        "type": "department",
        "department": "Pediatria",
        "startAt": "2026-06-12T09:00:00Z",
        "capacity": 12,
        "lecturer": "Mgr. Zuzana Farkašová",
        "location": "Pavilón D, miestnosť 2.14",
        "description": "Úvodné školenie pre nových členov pediatrického tímu.",
        "requirements": "Bez požiadaviek",
        "registrations": []
    }),
    training({
        "id": "radiology-mri-2026-06",
        "title": "MR bezpečnosť a kontrastné látky",
        "type": "specialization",
        "department": "Radiológia",
        "startAt": "2026-06-18T13:00:00Z",
        "capacity": 8,
        "lecturer": "MUDr. Peter Oravec",
        "location": "Radiológia, seminárna miestnosť",
        "description": "Špecializačné školenie pre bezpečnú prácu pri magnetickej rezonancii.",
        "requirements": "Platné školenie BOZP",
        "registrations": [
            registration("reg-mri-001", "EMP-3011", "MUDr. Ivana Bílá", "Radiológia", "2026-05-10T08:15:00Z"),
            registration("reg-mri-002", "EMP-3022", "Bc. Michal Rác", "Radiológia", "2026-05-10T08:35:00Z"),
            registration("reg-mri-003", "EMP-3050", "Mgr. Lenka Tóthová", "Radiológia", "2026-05-11T09:00:00Z"),
            registration("reg-mri-004", "EMP-3099", "MUDr. Daniel Kovář", "Urgent", "2026-05-12T15:20:00Z")
        ]
    }),
    training({
        "id": "anesthesia-airway-2026-06",
        "title": "Ťažká intubácia a krízové postupy",
        "type": "specialization",
        "department": "Anestéziológia",
        "startAt": "2026-06-25T07:30:00Z",
        "capacity": 3,
        "lecturer": "MUDr. Richard Urban",
        "location": "Simulačné centrum",
        "description": "Praktický nácvik postupov pri ťažkej intubácii.",
        "requirements": "Pracovné zaradenie na OAIM alebo urgentnom príjme",
        "registrations": [
            registration("reg-airway-001", "EMP-4012", "MUDr. Filip Šoltés", "Anestéziológia", "2026-05-09T07:50:00Z"),
            registration("reg-airway-002", "EMP-4028", "MUDr. Natália Marková", "Anestéziológia", "2026-05-09T08:10:00Z"),
            registration("reg-airway-003", "EMP-4055", "Bc. Roman Baláž", "Urgent", "2026-05-10T10:30:00Z")
        ]
    }),
    training({
        "id": "surgery-teamwork-2026-07",
        "title": "Simulačný tréning perioperačnej komunikácie",
        "type": "department",
        "department": "Chirurgia",
        "startAt": "2026-07-03T10:00:00Z",
        "capacity": 10,
        "lecturer": "MUDr. Samuel Konečný",
        "location": "Operačný trakt, tréningová sála",
        "description": "Tréning tímovej komunikácie počas perioperačných situácií.",
        "requirements": "Zaradenie v chirurgickom tíme",
        "status": "cancelled",
        "registrations": []
    }),
    training({
        "id": "gdpr-healthcare-2026-04",
        "title": "Ochrana osobných údajov v zdravotníctve",
        "type": "online",
        "department": "Interné",
        "startAt": "2026-04-15T08:30:00Z",
        "capacity": 15,
        "lecturer": "Mgr. Katarína Slámová",
        "onlineLink": "https://teams.example/mededu/gdpr-healthcare",
        "description": "Online školenie k práci s citlivými údajmi pacientov.",
        "requirements": "Prístup do nemocničného e-learningu",
        "status": "archived",
        "registrations": [
            registration("reg-gdpr-001", "EMP-5011", "Mgr. Hana Vršková", "Interné", "2026-04-01T08:00:00Z"),
            registration("reg-gdpr-002", "EMP-5026", "MUDr. Pavol Repa", "Interné", "2026-04-01T08:10:00Z"),
            registration("reg-gdpr-003", "EMP-5040", "Bc. Silvia Dúbravská", "Pediatria", "2026-04-02T09:30:00Z"),
            registration("reg-gdpr-004", "EMP-5062", "Mgr. Juraj Kováčik", "Chirurgia", "2026-04-03T11:45:00Z"),
            registration("reg-gdpr-005", "EMP-5078", "MUDr. Elena Gajdošová", "JIS", "2026-04-04T07:25:00Z"),
            registration("reg-gdpr-006", "EMP-5092", "Bc. Marek Gregor", "Radiológia", "2026-04-05T10:15:00Z"),
            registration("reg-gdpr-007", "EMP-5114", "Mgr. Veronika Sýkorová", "Urgent", "2026-04-06T13:05:00Z"),
            registration("reg-gdpr-008", "EMP-5121", "MUDr. Viktor Balog", "Anestéziológia", "2026-04-07T08:45:00Z"),
            registration("reg-gdpr-009", "EMP-5135", "Bc. Simona Pavlíková", "Interné", "2026-04-08T09:20:00Z"),
            registration("reg-gdpr-010", "EMP-5148", "Mgr. Karol Moravec", "Pediatria", "2026-04-09T15:30:00Z")
        ]
    })
]);

if (result.writeError) {
    console.error(result);
    print(`Error when writing the data: ${result.errmsg}`);
}

process.exit(0);
