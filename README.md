# Go Product Management System

A modern web application for product inventory management built with Go, featuring a beautiful blue and white theme with glass morphism design.

## Features

### 🔐 Authentication System
- User registration and login
- Role-based access control (Admin/User)
- Session management
- Beautiful notification system with auto-hide alerts

### 📦 Product Management
- Add new products with images
- Edit existing products
- Delete products (Admin only)
- Auto-increment ID system
- Image upload and management
- Product descriptions

### 🎨 Modern UI/UX
- Glass morphism design
- Blue and white color theme
- Responsive design for all devices
- Smooth animations and transitions
- Font Awesome icons
- Modern typography with Segoe UI

### 📱 Responsive Features
- Mobile-friendly interface
- Tablet optimization
- Desktop experience

## Tech Stack

- **Backend:** Go (Golang)
- **Database:** MySQL
- **Frontend:** HTML5, CSS3, JavaScript
- **Session Management:** Gorilla Sessions
- **Database Driver:** go-sql-driver/mysql
- **Icons:** Font Awesome 6

## Prerequisites

- Go 1.19 or higher
- MySQL 5.7 or higher
- Web browser with modern CSS support

## Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd Go
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up MySQL database**
   ```sql
   CREATE DATABASE db_go;
   USE db_go;
   
   -- Create users table
   CREATE TABLE users (
       id INT AUTO_INCREMENT PRIMARY KEY,
       username VARCHAR(50) UNIQUE NOT NULL,
       password VARCHAR(255) NOT NULL,
       role ENUM('admin', 'user') DEFAULT 'user'
   );
   
   -- Create products table
   CREATE TABLE good (
       id INT AUTO_INCREMENT PRIMARY KEY,
       name VARCHAR(255) NOT NULL,
       description TEXT,
       price DECIMAL(10,2) NOT NULL,
       image VARCHAR(255)
   );
   
   -- Insert default admin user (password: admin)
   INSERT INTO users (username, password, role) VALUES ('admin', 'admin', 'admin');
   ```

4. **Configure database connection**
   Update the database connection string in `main.go` if needed:
   ```go
   db, err = sql.Open("mysql", "root:@/db_go")
   ```

5. **Create uploads directory**
   ```bash
   mkdir uploads
   ```

6. **Run the application**
   ```bash
   go run main.go
   ```

7. **Access the application**
   Open your browser and navigate to `http://localhost:8080`

## Usage

### Default Login Credentials
- **Username:** admin
- **Password:** admin
- **Role:** Administrator

### User Roles

**Admin Users Can:**
- View all products
- Add new products
- Edit existing products
- Delete products
- Upload product images

**Regular Users Can:**
- View all products
- Browse product catalog

### Adding Products
1. Login as an admin user
2. Click "Add New Product"
3. Fill in product details:
   - Product name (required)
   - Description (optional)
   - Price (required)
   - Product image (optional)
4. Click "Add Product"

### Managing Products
- **Edit:** Click the edit icon next to any product
- **Delete:** Click the delete icon (confirmation required)
- **Images:** Upload new images or keep existing ones

## Project Structure

```
Go/
├── main.go              # Main application file
├── go.mod              # Go module file
├── go.sum              # Go dependencies
├── templates/          # HTML templates
│   ├── index.html      # Main product listing
│   ├── add.html        # Add product form
│   ├── edit.html       # Edit product form
│   ├── login.html      # Login/Register page
│   └── register.html   # Registration page
├── uploads/            # Uploaded product images
└── README.md          # This file
```

## API Endpoints

- `GET /` - Main product listing (requires login)
- `GET/POST /login` - User authentication
- `GET/POST /register` - User registration
- `GET/POST /add` - Add new product (admin only)
- `GET/POST /edit` - Edit product (admin only)
- `GET /delete` - Delete product (admin only)
- `GET /logout` - User logout
- `GET /uploads/*` - Static file serving for images

## Features in Detail

### Notification System
- Success notifications for login and registration
- Error notifications for failed operations
- Auto-hide success alerts after 3 seconds
- Smooth fade-out animations

### Image Management
- Automatic file naming with timestamps
- Support for common image formats
- Image preview in edit forms
- Automatic cleanup of old images when updating

### Security Features
- SQL injection prevention with prepared statements
- Session-based authentication
- Role-based access control
- Input validation and sanitization

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

This project is open source and available under the [MIT License](LICENSE).

## Screenshots

### Login Page
Modern authentication interface with glass morphism design.

### Product Dashboard
Clean and organized product listing with admin controls.

### Add/Edit Product
Intuitive forms for product management with image upload.

## Support

For support or questions, please open an issue in the GitHub repository.