#!/usr/bin/env node

const apiBase = (process.env.MEDEDU_API_BASE || "http://127.0.0.1:18080/api").replace(/\/$/, "");

function registration(employeeId, employeeName, department, extra = {}) {
  return {
    employeeId,
    employeeName,
    employeeEmail: extra.employeeEmail || `${employeeId.toLowerCase()}@hospital.example`,
    department,
    note: extra.note || "",
  };
}

const demoTrainings = [
  {
    title: "BOZP pre urgentný príjem",
    type: "mandatory",
    department: "Urgent",
    startAt: "2026-05-20T08:00:00Z",
    capacity: 20,
    lecturer: "Mgr. Jana Nováková",
    location: "Školiaca miestnosť A",
    onlineLink: "",
    description: "Povinné školenie bezpečnosti práce pre personál urgentného príjmu.",
    requirements: "Zamestnanecký preukaz",
    status: "planned",
    registrations: [
      registration("EMP-1042", "Bc. Peter Malina", "Urgent", { note: "Uprednostňuje ranný termín." }),
      registration("EMP-1077", "Mgr. Lucia Križová", "Chirurgia"),
      registration("EMP-1091", "MUDr. Martin Havel", "Urgent"),
      registration("EMP-1150", "Bc. Nina Bartošová", "Urgent"),
      registration("EMP-1216", "Mgr. Adam Kováč", "Interné"),
      registration("EMP-1320", "MUDr. Petra Malíková", "JIS"),
    ],
  },
  {
    title: "Prevencia infekcií na JIS",
    type: "department",
    department: "JIS",
    startAt: "2026-06-02T12:30:00Z",
    capacity: 2,
    lecturer: "MUDr. Eva Hrubá",
    location: "",
    onlineLink: "https://teams.example/mededu/icu-infection",
    description: "Praktické postupy prevencie nozokomiálnych infekcií a práca s izolačným režimom.",
    requirements: "Notebook alebo tablet",
    status: "planned",
    registrations: [
      registration("EMP-2018", "MUDr. Tomáš Benko", "JIS"),
      registration("EMP-2034", "Mgr. Andrea Poláková", "JIS"),
      registration("EMP-2071", "Bc. Katarína Švecová", "JIS", { note: "Môže prísť aj na náhradný termín." }),
      registration("EMP-2144", "MUDr. Jozef Varga", "Anestéziológia"),
    ],
  },
  {
    title: "Adaptačné školenie pre pediatriu",
    type: "department",
    department: "Pediatria",
    startAt: "2026-06-12T09:00:00Z",
    capacity: 12,
    lecturer: "Mgr. Zuzana Farkašová",
    location: "Pavilón D, miestnosť 2.14",
    onlineLink: "",
    description: "Úvodné školenie pre nových členov pediatrického tímu.",
    requirements: "Bez požiadaviek",
    status: "planned",
    registrations: [],
  },
  {
    title: "MR bezpečnosť a kontrastné látky",
    type: "specialization",
    department: "Radiológia",
    startAt: "2026-06-18T13:00:00Z",
    capacity: 8,
    lecturer: "MUDr. Peter Oravec",
    location: "Radiológia, seminárna miestnosť",
    onlineLink: "",
    description: "Špecializačné školenie pre bezpečnú prácu pri magnetickej rezonancii.",
    requirements: "Platné školenie BOZP",
    status: "planned",
    registrations: [
      registration("EMP-3011", "MUDr. Ivana Bílá", "Radiológia"),
      registration("EMP-3022", "Bc. Michal Rác", "Radiológia"),
      registration("EMP-3050", "Mgr. Lenka Tóthová", "Radiológia"),
      registration("EMP-3099", "MUDr. Daniel Kovář", "Urgent"),
    ],
  },
  {
    title: "Ťažká intubácia a krízové postupy",
    type: "specialization",
    department: "Anestéziológia",
    startAt: "2026-06-25T07:30:00Z",
    capacity: 3,
    lecturer: "MUDr. Richard Urban",
    location: "Simulačné centrum",
    onlineLink: "",
    description: "Praktický nácvik postupov pri ťažkej intubácii.",
    requirements: "Pracovné zaradenie na OAIM alebo urgentnom príjme",
    status: "planned",
    registrations: [
      registration("EMP-4012", "MUDr. Filip Šoltés", "Anestéziológia"),
      registration("EMP-4028", "MUDr. Natália Marková", "Anestéziológia"),
      registration("EMP-4055", "Bc. Roman Baláž", "Urgent"),
    ],
  },
  {
    title: "Simulačný tréning perioperačnej komunikácie",
    type: "department",
    department: "Chirurgia",
    startAt: "2026-07-03T10:00:00Z",
    capacity: 10,
    lecturer: "MUDr. Samuel Konečný",
    location: "Operačný trakt, tréningová sála",
    onlineLink: "",
    description: "Tréning tímovej komunikácie počas perioperačných situácií.",
    requirements: "Zaradenie v chirurgickom tíme",
    status: "cancelled",
    registrations: [],
  },
  {
    title: "Ochrana osobných údajov v zdravotníctve",
    type: "online",
    department: "Interné",
    startAt: "2026-04-15T08:30:00Z",
    capacity: 15,
    lecturer: "Mgr. Katarína Slámová",
    location: "",
    onlineLink: "https://teams.example/mededu/gdpr-healthcare",
    description: "Online školenie k práci s citlivými údajmi pacientov.",
    requirements: "Prístup do nemocničného e-learningu",
    status: "archived",
    registrations: [
      registration("EMP-5011", "Mgr. Hana Vršková", "Interné"),
      registration("EMP-5026", "MUDr. Pavol Repa", "Interné"),
      registration("EMP-5040", "Bc. Silvia Dúbravská", "Pediatria"),
      registration("EMP-5062", "Mgr. Juraj Kováčik", "Chirurgia"),
      registration("EMP-5078", "MUDr. Elena Gajdošová", "JIS"),
      registration("EMP-5092", "Bc. Marek Gregor", "Radiológia"),
      registration("EMP-5114", "Mgr. Veronika Sýkorová", "Urgent"),
      registration("EMP-5121", "MUDr. Viktor Balog", "Anestéziológia"),
      registration("EMP-5135", "Bc. Simona Pavlíková", "Interné"),
      registration("EMP-5148", "Mgr. Karol Moravec", "Pediatria"),
    ],
  },
];

async function request(path, options = {}) {
  const response = await fetch(`${apiBase}${path}`, {
    ...options,
    headers: {
      "content-type": "application/json",
      ...(options.headers || {}),
    },
  });

  if (!response.ok) {
    const body = await response.text();
    throw new Error(`${options.method || "GET"} ${path} failed: ${response.status} ${response.statusText} ${body}`);
  }

  if (response.status === 204) {
    return undefined;
  }

  return response.json();
}

async function seedDemoData() {
  console.log(`Seeding MedEdu demo data through ${apiBase}`);

  const existingTrainings = await request("/trainings");
  for (const training of existingTrainings) {
    await request(`/trainings/${encodeURIComponent(training.id)}`, { method: "DELETE" });
  }
  console.log(`Deleted ${existingTrainings.length} existing trainings`);

  for (const demoTraining of demoTrainings) {
    const { registrations, status, ...trainingInput } = demoTraining;
    const created = await request("/trainings", {
      method: "POST",
      body: JSON.stringify({ ...trainingInput, status: "planned" }),
    });

    for (const registrationInput of registrations) {
      await request(`/trainings/${encodeURIComponent(created.id)}/registrations`, {
        method: "POST",
        body: JSON.stringify(registrationInput),
      });
    }

    if (status !== "planned") {
      await request(`/trainings/${encodeURIComponent(created.id)}`, {
        method: "PUT",
        body: JSON.stringify({ ...trainingInput, status }),
      });
    }

    console.log(`Created ${demoTraining.title} (${registrations.length} registrations, ${status})`);
  }

  console.log("Demo data ready");
}

seedDemoData().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
