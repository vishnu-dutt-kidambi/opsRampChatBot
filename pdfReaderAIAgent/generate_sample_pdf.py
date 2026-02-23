"""Generate a sample PDF for testing the PDF Q&A Agent."""
from fpdf import FPDF

pdf = FPDF()
pdf.set_auto_page_break(auto=True, margin=15)

# --- Page 1: Title + Intro ---
pdf.add_page()
pdf.set_font("Helvetica", "B", 22)
pdf.cell(0, 15, "The Solar System: A Quick Guide", new_x="LMARGIN", new_y="NEXT", align="C")
pdf.ln(10)

pdf.set_font("Helvetica", "", 12)
intro = (
    "The Solar System is the gravitationally bound system of the Sun and the objects that orbit it. "
    "It formed approximately 4.6 billion years ago from the gravitational collapse of a giant "
    "interstellar molecular cloud. The vast majority of the system's mass is in the Sun, with most "
    "of the remaining mass contained in the planet Jupiter. The four inner planets - Mercury, Venus, "
    "Earth, and Mars - are called terrestrial planets because they have rocky surfaces. The four outer "
    "planets - Jupiter, Saturn, Uranus, and Neptune - are called gas giants (or ice giants for Uranus "
    "and Neptune) because they are much larger and composed primarily of hydrogen and helium."
)
pdf.multi_cell(0, 7, intro)
pdf.ln(5)

# --- Section: The Sun ---
pdf.set_font("Helvetica", "B", 16)
pdf.cell(0, 10, "The Sun", new_x="LMARGIN", new_y="NEXT")
pdf.set_font("Helvetica", "", 12)
sun_text = (
    "The Sun is the star at the center of the Solar System. It is a nearly perfect ball of hot plasma, "
    "with a diameter of about 1,391,000 kilometers (864,000 miles), which is approximately 109 times "
    "the diameter of Earth. The Sun's mass is about 330,000 times that of Earth and accounts for "
    "roughly 99.86% of the total mass of the Solar System. The Sun's core temperature is approximately "
    "15 million degrees Celsius (27 million degrees Fahrenheit). The Sun generates energy through "
    "nuclear fusion, converting hydrogen into helium at a rate of about 600 million tons of hydrogen "
    "per second. The Sun is classified as a G-type main-sequence star (G2V), informally known as a "
    "yellow dwarf. The Sun's age is estimated to be about 4.6 billion years, and it is expected to "
    "remain on the main sequence for another 5 billion years before expanding into a red giant."
)
pdf.multi_cell(0, 7, sun_text)
pdf.ln(5)

# --- Section: Inner Planets ---
pdf.set_font("Helvetica", "B", 16)
pdf.cell(0, 10, "The Inner Planets (Terrestrial Planets)", new_x="LMARGIN", new_y="NEXT")

planets_inner = [
    ("Mercury", (
        "Mercury is the smallest planet in the Solar System and the closest to the Sun. It has a "
        "diameter of 4,879 km and orbits the Sun once every 88 Earth days. Mercury has no atmosphere "
        "to retain heat, so temperatures range from -180 degrees C at night to 430 degrees C during the day. "
        "Its surface is heavily cratered, similar to the Moon. Mercury has no moons and no rings."
    )),
    ("Venus", (
        "Venus is the second planet from the Sun and is similar in size to Earth, with a diameter of "
        "12,104 km. It has a thick, toxic atmosphere composed primarily of carbon dioxide with clouds "
        "of sulfuric acid, creating a runaway greenhouse effect. The surface temperature averages "
        "465 degrees C, making it the hottest planet in the Solar System - even hotter than Mercury. "
        "Venus rotates very slowly and in the opposite direction of most planets. A day on Venus "
        "(243 Earth days) is longer than a year on Venus (225 Earth days). Venus has no moons."
    )),
    ("Earth", (
        "Earth is the third planet from the Sun and the only known planet to support life. It has "
        "a diameter of 12,742 km and is approximately 150 million km from the Sun (1 Astronomical "
        "Unit). Earth's atmosphere is composed of 78% nitrogen, 21% oxygen, and 1% other gases. "
        "About 71% of Earth's surface is covered by water. Earth has one natural satellite, the Moon, "
        "which has a diameter of 3,474 km. The Earth's axial tilt of 23.5 degrees is responsible for "
        "the seasons. Earth's magnetic field protects it from harmful solar radiation."
    )),
    ("Mars", (
        "Mars is the fourth planet from the Sun, known as the Red Planet due to iron oxide (rust) "
        "on its surface. It has a diameter of 6,779 km, roughly half the size of Earth. Mars has a "
        "thin atmosphere composed mostly of carbon dioxide. The planet features Olympus Mons, the "
        "tallest volcano in the Solar System at about 21.9 km high, and Valles Marineris, a canyon "
        "system stretching over 4,000 km. Mars has two small moons: Phobos and Deimos. Surface "
        "temperatures on Mars average about -60 degrees C. Scientists have found evidence of ancient "
        "river beds and ice caps, suggesting water once flowed on its surface."
    )),
]

for name, desc in planets_inner:
    pdf.set_font("Helvetica", "BI", 13)
    pdf.cell(0, 8, name, new_x="LMARGIN", new_y="NEXT")
    pdf.set_font("Helvetica", "", 12)
    pdf.multi_cell(0, 7, desc)
    pdf.ln(3)

# --- Page break, Outer Planets ---
pdf.add_page()
pdf.set_font("Helvetica", "B", 16)
pdf.cell(0, 10, "The Outer Planets (Gas and Ice Giants)", new_x="LMARGIN", new_y="NEXT")

planets_outer = [
    ("Jupiter", (
        "Jupiter is the fifth planet from the Sun and the largest planet in the Solar System. Its "
        "diameter is 139,820 km - more than 11 times that of Earth. Jupiter is a gas giant composed "
        "mainly of hydrogen and helium. Its most famous feature is the Great Red Spot, a massive "
        "storm that has been raging for at least 400 years and is larger than Earth. Jupiter has at "
        "least 95 known moons, including the four large Galilean moons: Io, Europa, Ganymede, and "
        "Callisto. Europa is of particular interest because scientists believe it has a subsurface "
        "ocean that could potentially harbor life. Jupiter's powerful magnetic field is the strongest "
        "of any planet. One year on Jupiter equals about 11.86 Earth years."
    )),
    ("Saturn", (
        "Saturn is the sixth planet from the Sun and is best known for its spectacular ring system. "
        "It has a diameter of 116,460 km, making it the second-largest planet. Saturn is a gas giant "
        "composed mostly of hydrogen and helium, and it is the least dense planet - it would float in "
        "water if a large enough body of water existed. Saturn's rings are made of billions of particles "
        "of ice and rock, ranging in size from tiny grains to chunks as large as houses. The rings "
        "extend up to 282,000 km from the planet. Saturn has at least 146 known moons. Its largest moon, "
        "Titan, is the second-largest moon in the Solar System and has a thick nitrogen atmosphere "
        "with methane lakes on its surface."
    )),
    ("Uranus", (
        "Uranus is the seventh planet from the Sun and is classified as an ice giant. It has a diameter "
        "of 50,724 km and is composed primarily of water, methane, and ammonia ices surrounding a small "
        "rocky core. Uranus is unique because it rotates on its side, with an axial tilt of about 98 "
        "degrees - likely the result of a collision with an Earth-sized object long ago. This extreme "
        "tilt means its poles take turns pointing directly at the Sun. Uranus appears pale blue-green "
        "due to methane in its atmosphere. It has 13 known rings and 28 known moons. One year on "
        "Uranus equals about 84 Earth years. Uranus was the first planet discovered using a telescope, "
        "found by William Herschel in 1781."
    )),
    ("Neptune", (
        "Neptune is the eighth and farthest planet from the Sun in the Solar System. It has a diameter "
        "of 49,244 km and, like Uranus, is classified as an ice giant. Neptune is known for having the "
        "strongest winds in the Solar System, reaching speeds of up to 2,100 km/h. The planet appears "
        "deep blue due to methane in its atmosphere. Neptune has 16 known moons, the largest being "
        "Triton, which orbits in the opposite direction of Neptune's rotation - suggesting it was "
        "captured from the Kuiper Belt. Neptune has 5 known rings. One year on Neptune equals about "
        "164.8 Earth years. Neptune was the first planet to be found through mathematical prediction "
        "rather than direct observation, discovered in 1846."
    )),
]

for name, desc in planets_outer:
    pdf.set_font("Helvetica", "BI", 13)
    pdf.cell(0, 8, name, new_x="LMARGIN", new_y="NEXT")
    pdf.set_font("Helvetica", "", 12)
    pdf.multi_cell(0, 7, desc)
    pdf.ln(3)

# --- Page 3: Dwarf Planets & Fun Facts ---
pdf.add_page()
pdf.set_font("Helvetica", "B", 16)
pdf.cell(0, 10, "Dwarf Planets", new_x="LMARGIN", new_y="NEXT")
pdf.set_font("Helvetica", "", 12)
dwarf = (
    "The Solar System also contains five recognized dwarf planets: Pluto, Eris, Haumea, Makemake, "
    "and Ceres. Pluto was considered the ninth planet until 2006, when the International Astronomical "
    "Union (IAU) reclassified it as a dwarf planet. Pluto has a diameter of 2,377 km and has five "
    "known moons, the largest being Charon. Ceres is the only dwarf planet located in the inner "
    "Solar System, residing in the asteroid belt between Mars and Jupiter. It has a diameter of "
    "about 940 km and was the first asteroid ever discovered in 1801. Eris is slightly smaller than "
    "Pluto but more massive, and its discovery in 2005 was a key factor in the redefinition of planet."
)
pdf.multi_cell(0, 7, dwarf)
pdf.ln(5)

pdf.set_font("Helvetica", "B", 16)
pdf.cell(0, 10, "Key Facts and Numbers", new_x="LMARGIN", new_y="NEXT")
pdf.set_font("Helvetica", "", 12)

facts = [
    "The speed of light is approximately 299,792 km/s. Light from the Sun reaches Earth in about 8 minutes and 20 seconds.",
    "The asteroid belt between Mars and Jupiter contains millions of rocky objects, but its total mass is less than 4% of the Moon's mass.",
    "The Kuiper Belt extends from Neptune's orbit (30 AU) to approximately 50 AU from the Sun and contains icy bodies including Pluto.",
    "The Oort Cloud is a theoretical spherical shell of icy objects surrounding the Solar System at distances of 2,000 to 200,000 AU.",
    "Voyager 1, launched in 1977, is the most distant human-made object, currently over 24 billion km from Earth.",
    "The Solar System orbits the center of the Milky Way galaxy at about 828,000 km/h, completing one orbit every 225-250 million years.",
    "The total number of known moons in the Solar System exceeds 290 as of 2024.",
    "Jupiter's moon Ganymede is the largest moon in the Solar System, with a diameter of 5,268 km - larger than Mercury.",
]

for fact in facts:
    pdf.cell(5, 7, "-")
    pdf.multi_cell(0, 7, "  " + fact)
    pdf.ln(2)

# --- Save ---
output_path = "pdfs/solar_system_guide.pdf"
pdf.output(output_path)
print("Sample PDF created: " + output_path)
print("Pages: " + str(pdf.pages_count))
